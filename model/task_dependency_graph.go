package model

import (
	"github.com/evergreen-ci/evergreen/model/task"
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/multi"
	"gonum.org/v1/gonum/graph/topo"
	"gonum.org/v1/gonum/graph/traverse"
)

type taskDependencyGraph struct {
	graph        *multi.DirectedGraph
	tasksToNodes map[TVPair]graph.Node
	nodesToTasks map[graph.Node]TVPair
	edgesToDeps  map[graph.Edge]TaskUnitDependency
}

func ProjectDependencyGraph(p *Project) taskDependencyGraph {
	g := newDependencyGraph()
	g.buildFromProject(p)

	return g
}

func versionDependencyGraph(versionID string) (taskDependencyGraph, error) {
	g := newDependencyGraph()
	tasks, err := task.FindAllTasksFromVersionWithDependencies(versionID)
	if err != nil {
		return g, errors.Wrapf(err, "getting tasks for version '%s'", versionID)
	}

	g.buildFromTasks(tasks)
	return g, nil
}

func newDependencyGraph() taskDependencyGraph {
	return taskDependencyGraph{
		graph:        multi.NewDirectedGraph(),
		tasksToNodes: make(map[TVPair]graph.Node),
		nodesToTasks: make(map[graph.Node]TVPair),
		edgesToDeps:  make(map[graph.Edge]TaskUnitDependency),
	}
}

func (g *taskDependencyGraph) buildFromTasks(tasks []task.Task) error {
	taskIDToTV := make(map[string]TVPair)
	for _, task := range tasks {
		taskTV := TVPair{TaskName: task.DisplayName, Variant: task.BuildVariant}
		g.addTaskNode(taskTV)
		taskIDToTV[task.Id] = taskTV
	}

	for _, task := range tasks {
		dependentTaskTV := TVPair{TaskName: task.DisplayName, Variant: task.BuildVariant}
		for _, dep := range task.DependsOn {
			dependedOnTaskTV := taskIDToTV[dep.TaskId]
			g.addEdge(dependentTaskTV, dependedOnTaskTV, TaskUnitDependency{
				Name:    dependedOnTaskTV.TaskName,
				Variant: dependedOnTaskTV.Variant,
				Status:  dep.Status,
			})
		}
	}

	return nil
}

func (g *taskDependencyGraph) buildFromProject(p *Project) {
	tasks := p.FindAllBuildVariantTasks()
	var taskTVs []TVPair

	for _, task := range tasks {
		g.addTaskNode(task.ToTVPair())
		taskTVs = append(taskTVs, task.ToTVPair())
	}

	for _, task := range tasks {
		for dep, depTasks := range dependenciesForTaskUnit(task, taskTVs) {
			for _, depTask := range depTasks {
				g.addEdge(task.ToTVPair(), depTask, dep)
			}
		}
	}
}

// dependenciesForTaskUnit returns a map of dependencies to the task variant pairs they match.
func dependenciesForTaskUnit(dependentTaskUnit BuildVariantTaskUnit, allTVPairs []TVPair) map[TaskUnitDependency][]TVPair {
	dependencies := make(map[TaskUnitDependency][]TVPair)
	for _, dep := range dependentTaskUnit.DependsOn {
		// Use the current variant if none is specified.
		if dep.Variant == "" {
			dep.Variant = dependentTaskUnit.Variant
		}

		for _, tv := range allTVPairs {
			depTV := TVPair{Variant: tv.Variant, TaskName: tv.TaskName}
			if depTV != dependentTaskUnit.ToTVPair() &&
				(dep.Variant == AllVariants || depTV.Variant == dep.Variant) &&
				(dep.Name == AllDependencies || depTV.TaskName == dep.Name) {
				dependencies[dep] = append(dependencies[dep], depTV)
			}
		}
	}

	return dependencies
}

func (g *taskDependencyGraph) addTaskNode(task TVPair) {
	node := g.graph.NewNode()
	g.graph.AddNode(node)
	g.tasksToNodes[task] = node
	g.nodesToTasks[node] = task
}

func (g *taskDependencyGraph) addEdge(dependentTask, dependedOnTask TVPair, dep TaskUnitDependency) {
	dependentNode := g.tasksToNodes[dependentTask]
	dependedOnNode := g.tasksToNodes[dependedOnTask]
	if dependentNode == nil || dependedOnNode == nil {
		return
	}

	line := g.graph.NewLine(dependentNode, dependedOnNode)
	g.graph.SetLine(line)
	g.edgesToDeps[g.graph.Edge(g.tasksToNodes[dependentTask].ID(), g.tasksToNodes[dependedOnTask].ID())] = dep
}

func (g *taskDependencyGraph) tasksDependingOnTask(t TVPair) []TVPair {
	dependedOnNode := g.tasksToNodes[t]
	if dependedOnNode == nil {
		return nil
	}

	var dependentTasks []TVPair
	nodes := g.graph.To(dependedOnNode.ID())
	for nodes.Next() {
		dependentTasks = append(dependentTasks, g.nodesToTasks[nodes.Node()])
	}

	return dependentTasks
}

func (g *taskDependencyGraph) getDependencyEdge(dependentTask, dependedOnTask TVPair) (TaskUnitDependency, error) {
	edge := g.graph.Edge(g.tasksToNodes[dependentTask].ID(), g.tasksToNodes[dependedOnTask].ID())
	if edge == nil {
		return TaskUnitDependency{}, errors.Errorf("'%s' has no dependency on '%s'", dependentTask, dependedOnTask)
	}

	return g.edgesToDeps[edge], nil
}

func (g *taskDependencyGraph) Cycles() [][]TVPair {
	var cycles [][]TVPair
	stronglyConnectedComponenets := topo.TarjanSCC(g.graph)
	for _, scc := range stronglyConnectedComponenets {
		if len(scc) <= 1 {
			continue
		}

		var cycle []TVPair
		for _, node := range scc {
			taskInCycle := g.nodesToTasks[node]
			cycle = append(cycle, taskInCycle)
		}
		cycles = append(cycles, cycle)
	}

	return cycles
}

func (g *taskDependencyGraph) DepthFirstSearch(start, target TVPair, traverseEdge func(dependentTask, dependedOnTask TVPair) bool) bool {
	traversal := traverse.DepthFirst{
		Traverse: func(e graph.Edge) bool {
			if traverseEdge == nil {
				return true
			}

			dependedOn := g.nodesToTasks[e.From()]
			dependent := g.nodesToTasks[e.To()]

			return traverseEdge(dependent, dependedOn)
		},
	}

	return traversal.Walk(g.graph, g.tasksToNodes[start], func(n graph.Node) bool { return g.nodesToTasks[n] == target }) == nil
}
