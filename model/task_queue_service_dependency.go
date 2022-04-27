package model

import (
	"sort"
	"sync"
	"time"

	"github.com/evergreen-ci/evergreen"
	"github.com/evergreen-ci/evergreen/model/host"
	"github.com/evergreen-ci/evergreen/model/task"
	"github.com/evergreen-ci/utility"
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/message"
	"github.com/pkg/errors"
)

const (
	DAGDispatcher = "DAG-task-dispatcher"
)

type basicCachedDAGDispatcherImpl struct {
	mu            sync.RWMutex
	distroID      string
	taskQueue     []TaskQueueItem
	queueIndexMap map[string]int
	taskGroups    map[string]schedulableUnit // map[compositeGroupID(TaskQueueItem.Group, TaskQueueItem.BuildVariant, TaskQueueItem.Project, TaskQueueItem.Version)]schedulableUnit
	ttl           time.Duration
	lastUpdated   time.Time
}

// newDistroTaskDAGDispatchService creates a basicCachedDAGDispatcherImpl from a slice of TaskQueueItems.
func newDistroTaskDAGDispatchService(taskQueue TaskQueue, ttl time.Duration) (*basicCachedDAGDispatcherImpl, error) {
	d := &basicCachedDAGDispatcherImpl{
		distroID: taskQueue.Distro,
		ttl:      ttl,
	}

	if taskQueue.Length() != 0 {
		if err := d.rebuild(taskQueue.Queue); err != nil {
			return nil, errors.Wrapf(err, "creating distro DAG task dispatch service for distro '%s'", taskQueue.Distro)
		}
	}

	grip.Debug(message.Fields{
		"dispatcher":         DAGDispatcher,
		"function":           "newDistroTaskDAGDispatchService",
		"message":            "initializing new basicCachedDAGDispatcherImpl for a distro",
		"distro_id":          d.distroID,
		"ttl":                d.ttl,
		"last_updated":       d.lastUpdated,
		"num_task_groups":    len(d.taskGroups),
		"num_taskqueueitems": taskQueue.Length(),
	})

	return d, nil
}

func (d *basicCachedDAGDispatcherImpl) Type() string {
	return evergreen.DispatcherVersionRevisedWithDependencies
}

func (d *basicCachedDAGDispatcherImpl) CreatedAt() time.Time {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.lastUpdated
}

func (d *basicCachedDAGDispatcherImpl) Refresh() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !shouldRefreshCached(d.ttl, d.lastUpdated) {
		return nil
	}

	taskQueue, err := FindDistroTaskQueue(d.distroID)
	if err != nil {
		return errors.WithStack(err)
	}

	taskQueueItems := taskQueue.Queue
	if err := d.rebuild(taskQueueItems); err != nil {
		return errors.Wrapf(err, "building the directed graph for distro '%s'", d.distroID)
	}

	grip.Debug(message.Fields{
		"dispatcher":                 DAGDispatcher,
		"function":                   "Refresh",
		"message":                    "refresh was successful",
		"distro_id":                  d.distroID,
		"num_task_groups":            len(d.taskGroups),
		"initial_num_taskqueueitems": len(taskQueueItems),
		"sorted_num_taskqueueitems":  len(d.taskQueue),
		"refreshed_at:":              time.Now(),
	})

	return nil
}

func (d *basicCachedDAGDispatcherImpl) rebuild(items []TaskQueueItem) error {
	d.taskGroups = map[string]schedulableUnit{} // map[compositeGroupID(TaskQueueItem.Group, TaskQueueItem.BuildVariant, TaskQueueItem.Project, TaskQueueItem.Version)]schedulableUnit
	for _, item := range items {
		if item.Group != "" {
			// If it's the first time encountering the task group create an entry for it in the taskGroups map.
			// Otherwise, append to the taskQueueItem array in the map.
			id := compositeGroupID(item.Group, item.BuildVariant, item.Project, item.Version)
			if _, ok := d.taskGroups[id]; !ok {
				d.taskGroups[id] = schedulableUnit{
					id:       id,
					group:    item.Group,
					project:  item.Project,
					version:  item.Version,
					variant:  item.BuildVariant,
					maxHosts: item.GroupMaxHosts,
					tasks:    []TaskQueueItem{item},
				}
			} else {
				taskGroup := d.taskGroups[id]
				taskGroup.tasks = append(taskGroup.tasks, item)
				d.taskGroups[id] = taskGroup
			}
		}
	}

	// Reorder the schedulableUnit.tasks by taskQueueItem.GroupIndex.
	// For a single host task group (MaxHosts: 1) this ensures that its tasks are dispatched in the desired order.
	for _, su := range d.taskGroups {
		sort.SliceStable(su.tasks, func(i, j int) bool { return su.tasks[i].GroupIndex < su.tasks[j].GroupIndex })
	}

	if err := d.makeQueue(items); err != nil {
		return errors.Wrap(err, "sorting items in the queue")
	}
	for i, item := range d.taskQueue {
		d.queueIndexMap[item.Id] = i
	}

	d.lastUpdated = time.Now()

	return nil
}

func (d *basicCachedDAGDispatcherImpl) makeQueue(items []TaskQueueItem) error {
	g := task.NewDependencyGraph()
	nodeItemMap := map[task.TaskNode]TaskQueueItem{}
	IDNodeMap := map[string]task.TaskNode{}

	for _, item := range items {
		g.AddTaskNode(item.ToTaskNode())
		nodeItemMap[item.ToTaskNode()] = item
		IDNodeMap[item.Id] = item.ToTaskNode()
	}

	for _, item := range items {
		for _, dependency := range item.Dependencies {
			g.AddReverseEdge(item.ToTaskNode(), IDNodeMap[dependency], task.DependencyEdge{})
		}
	}

	sortedNodes, err := g.TopologicalStableSort()
	if err != nil {
		return errors.Wrap(err, "sorting nodes in the graph")
	}

	if cycles := g.Cycles(); len(cycles) > 0 {
		grip.Error(message.Fields{
			"dispatcher": DAGDispatcher,
			"function":   "rebuild",
			"message":    "tasks in the queue form dependency cycle(s)",
			"cycles":     cycles,
			"distro_id":  d.distroID,
		})
	}

	d.taskQueue = make([]TaskQueueItem, 0, len(sortedNodes))
	for _, node := range sortedNodes {
		d.taskQueue = append(d.taskQueue, nodeItemMap[node])
	}

	return nil
}

// FindNextTask returns the next dispatchable task in the queue, and returns the tasks that need to be checked for dependencies.
func (d *basicCachedDAGDispatcherImpl) FindNextTask(spec TaskSpec, amiUpdatedTime time.Time) *TaskQueueItem {
	d.mu.Lock()
	defer d.mu.Unlock()
	// If the host just ran a task group, give it one back.
	if spec.Group != "" {
		taskGroupID := compositeGroupID(spec.Group, spec.BuildVariant, spec.Project, spec.Version)
		taskGroupUnit, ok := d.taskGroups[taskGroupID] // schedulableUnit
		if ok {
			if next := d.nextTaskGroupTask(taskGroupUnit); next != nil {
				// Mark the task as dispatched in the task queue.
				d.taskQueue[d.queueIndexMap[next.Id]].IsDispatched = true
				return next
			}
		}
		// If the task group is not present in the TaskGroups map, then all its tasks are considered dispatched.
		// Fall through to get a task that's not in this task group.

		grip.Debug(message.Fields{
			"dispatcher":               DAGDispatcher,
			"function":                 "FindNextTask",
			"message":                  "basicCachedDAGDispatcherImpl.taskGroupTasks[key] was not found - assuming it has been dispatched; falling through to try and get a task not in the current task group",
			"key":                      taskGroupID,
			"taskspec_group":           spec.Group,
			"taskspec_build_variant":   spec.BuildVariant,
			"taskspec_version":         spec.Version,
			"taskspec_project":         spec.Project,
			"taskspec_group_max_hosts": spec.GroupMaxHosts,
			"distro_id":                d.distroID,
		})
	}
	var numIterated int
	dependencyCaches := make(map[string]task.Task)
	for i, item := range d.taskQueue {
		numIterated += 1

		// TODO Consider checking if the state of any task has changed, which could unblock later tasks in the queue.
		// Currently, we just wait for the dispatcher's in-memory queue to refresh.

		// If maxHosts is not set, this is not a task group.
		if item.GroupMaxHosts == 0 {
			// Dispatch this standalone task if all of the following are true:
			// (a) it's not marked as dispatched in the in-memory queue.
			// (b) a record of the task exists in the database.
			// (c) it never previously ran on another host.
			// (d) all of its dependencies are satisfied.

			if item.IsDispatched {
				continue
			}

			nextTaskFromDB, err := task.FindOneId(item.Id)
			if err != nil {
				grip.Warning(message.WrapError(err, message.Fields{
					"dispatcher": DAGDispatcher,
					"function":   "FindNextTask",
					"message":    "problem finding task in db",
					"task_id":    item.Id,
					"distro_id":  d.distroID,
				}))
				return nil
			}
			if nextTaskFromDB == nil {
				grip.Warning(message.Fields{
					"dispatcher": DAGDispatcher,
					"function":   "FindNextTask",
					"message":    "task from db not found",
					"task_id":    item.Id,
					"distro_id":  d.distroID,
				})
				return nil
			}

			// Cache the task as dispatched from the in-memory queue's point of view.
			// However, it won't actually be dispatched to a host if it doesn't satisfy all constraints.
			d.taskQueue[i].IsDispatched = true

			if nextTaskFromDB.StartTime != utility.ZeroTime {
				continue
			}

			dependenciesMet, err := nextTaskFromDB.DependenciesMet(dependencyCaches)
			if err != nil {
				grip.Warning(message.WrapError(err, message.Fields{
					"dispatcher": DAGDispatcher,
					"function":   "FindNextTask",
					"message":    "error checking dependencies for task",
					"outcome":    "skip and continue",
					"task":       item.Id,
					"distro_id":  d.distroID,
				}))
				continue
			}

			if !dependenciesMet {
				continue
			}

			// AMI Updated time is only provided if the host is running with an outdated AMI.
			// If the task was created after the time that the AMI was updated, then we should wait for an updated host.
			if !utility.IsZeroTime(amiUpdatedTime) && nextTaskFromDB.IngestTime.After(amiUpdatedTime) {
				grip.Debug(message.Fields{
					"dispatcher":       DAGDispatcher,
					"function":         "FindNextTask",
					"message":          "skipping because AMI is outdated",
					"task_id":          nextTaskFromDB.Id,
					"distro_id":        d.distroID,
					"ami_updated_time": amiUpdatedTime,
					"ingest_time":      nextTaskFromDB.IngestTime,
				})
				continue
			}
			return &item
		}

		// For a task group task, do some arithmetic to see if the group's next task is dispatchable.
		taskGroupID := compositeGroupID(item.Group, item.BuildVariant, item.Project, item.Version)
		taskGroupUnit, ok := d.taskGroups[taskGroupID]
		if !ok {
			continue
		}

		if taskGroupUnit.runningHosts < taskGroupUnit.maxHosts {
			numHosts, err := host.NumHostsByTaskSpec(item.Group, item.BuildVariant, item.Project, item.Version)
			if err != nil {
				grip.Warning(message.WrapError(err, message.Fields{
					"dispatcher": DAGDispatcher,
					"function":   "FindNextTask",
					"message":    "problem running NumHostsByTaskSpec query - returning nil",
					"group":      item.Group,
					"variant":    item.BuildVariant,
					"project":    item.Project,
					"version":    item.Version,
					"distro_id":  d.distroID,
				}))
				return nil
			}

			taskGroupUnit.runningHosts = numHosts
			d.taskGroups[taskGroupID] = taskGroupUnit
			if taskGroupUnit.runningHosts < taskGroupUnit.maxHosts {
				if next := d.nextTaskGroupTask(taskGroupUnit); next != nil {
					// Mark the task as dispatched in the task queue.
					d.taskQueue[d.queueIndexMap[next.Id]].IsDispatched = true
					return next
				}
			}
		}
	}
	return nil
}

func (d *basicCachedDAGDispatcherImpl) nextTaskGroupTask(unit schedulableUnit) *TaskQueueItem {
	for i, nextTaskQueueItem := range unit.tasks {
		// Dispatch this task if all of the following are true:
		// (a) it's not marked as dispatched in the in-memory queue.
		// (b) a record of the task exists in the database.
		// (c) if it belongs to a TaskGroup bound to a single host - it's not blocked by a previous task within the TaskGroup that failed.
		// (d) it never previously ran on another host.
		// (e) all of its dependencies are satisfied.

		if nextTaskQueueItem.IsDispatched == true {
			continue
		}

		nextTaskFromDB, err := task.FindOneId(nextTaskQueueItem.Id)
		if err != nil {
			grip.Warning(message.WrapError(err, message.Fields{
				"dispatcher": DAGDispatcher,
				"function":   "nextTaskGroupTask",
				"message":    "problem finding task in db",
				"task_id":    nextTaskQueueItem.Id,
				"distro_id":  d.distroID,
			}))
			return nil
		}
		if nextTaskFromDB == nil {
			grip.Warning(message.Fields{
				"dispatcher": DAGDispatcher,
				"function":   "nextTaskGroupTask",
				"message":    "task from db not found",
				"task_id":    nextTaskQueueItem.Id,
				"distro_id":  d.distroID,
			})
			return nil
		}

		// Cache the task as dispatched from the in-memory queue's point of view.
		// However, it won't actually be dispatched to a host if it doesn't satisfy all constraints.
		d.taskGroups[unit.id].tasks[i].IsDispatched = true

		if isBlockedSingleHostTaskGroup(unit, nextTaskFromDB) {
			delete(d.taskGroups, unit.id)
			return nil
		}

		if nextTaskFromDB.StartTime != utility.ZeroTime {
			continue
		}

		dependencyCaches := make(map[string]task.Task)
		dependenciesMet, err := nextTaskFromDB.DependenciesMet(dependencyCaches)
		if err != nil {
			grip.Warning(message.WrapError(err, message.Fields{
				"dispatcher": DAGDispatcher,
				"function":   "nextTaskGroupTask",
				"message":    "error checking dependencies for task",
				"outcome":    "skip and continue",
				"task_id":    nextTaskQueueItem.Id,
				"distro_id":  d.distroID,
			}))
			continue
		}

		if !dependenciesMet {
			continue
		}

		// If this is the last task in the schedulableUnit.tasks, delete the task group.
		if i == len(unit.tasks)-1 {
			delete(d.taskGroups, unit.id)
		}

		return &nextTaskQueueItem
	}

	return nil
}
