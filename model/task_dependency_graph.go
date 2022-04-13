package model

import (
	"github.com/evergreen-ci/evergreen/model/task"
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/multi"
	"gonum.org/v1/gonum/graph/topo"
)

type taskDependencyGraph struct {
	Graph        *multi.DirectedGraph
	TasksToNodes map[TVPair]graph.Node
	NodesToTasks map[graph.Node]TVPair
	EdgesToDeps  map[graph.Edge]TaskUnitDependency
}

func NewDependencyGraphFromProject(p *Project) taskDependencyGraph {
	g := newTaskDependencyGraph()
	g.buildFromProject(p)

	return g
}

func NewDependencyGraphFromVersion(versionID string) (taskDependencyGraph, error) {
	g := newTaskDependencyGraph()
	return g, g.buildFromVersion(versionID)
}

func newTaskDependencyGraph() taskDependencyGraph {
	return taskDependencyGraph{
		Graph:        multi.NewDirectedGraph(),
		TasksToNodes: make(map[TVPair]graph.Node),
		NodesToTasks: make(map[graph.Node]TVPair),
		EdgesToDeps:  make(map[graph.Edge]TaskUnitDependency),
	}
}

func (g *taskDependencyGraph) buildFromVersion(versionID string) error {
	tasks, err := task.FindAllTasksFromVersionWithDependencies(versionID)
	if err != nil {
		return errors.Wrapf(err, "getting tasks for version '%s'", versionID)
	}

	taskIDToTV := make(map[string]TVPair)
	for _, task := range tasks {
		node := g.Graph.NewNode()
		g.Graph.AddNode(node)

		taskTV := TVPair{TaskName: task.DisplayName, Variant: task.BuildVariant}
		g.TasksToNodes[taskTV] = node
		g.NodesToTasks[node] = taskTV
		taskIDToTV[task.Id] = taskTV
	}

	for _, task := range tasks {
		dependentTaskTV := TVPair{TaskName: task.DisplayName, Variant: task.BuildVariant}
		for _, dep := range task.DependsOn {
			dependedOnTaskTV := taskIDToTV[dep.TaskId]
			line := g.Graph.NewLine(g.TasksToNodes[dependedOnTaskTV], g.TasksToNodes[dependentTaskTV])
			g.Graph.SetLine(line)

			g.EdgesToDeps[g.Graph.Edge(g.TasksToNodes[dependedOnTaskTV].ID(), g.TasksToNodes[dependentTaskTV].ID())] = TaskUnitDependency{
				Name:    dependedOnTaskTV.TaskName,
				Variant: dependedOnTaskTV.Variant,
				Status:  dep.Status,
			}
		}
	}

	return nil
}

func (g *taskDependencyGraph) buildFromProject(p *Project) {
	tasks := p.FindAllBuildVariantTasks()
	var taskTVs []TVPair

	for _, task := range tasks {
		node := g.Graph.NewNode()
		g.Graph.AddNode(node)
		g.TasksToNodes[task.ToTVPair()] = node
		g.NodesToTasks[node] = task.ToTVPair()

		taskTVs = append(taskTVs, task.ToTVPair())
	}

	for _, task := range tasks {
		for dep, depTasks := range dependenciesForTaskUnit(task, taskTVs) {
			for _, depTask := range depTasks {
				line := g.Graph.NewLine(g.TasksToNodes[task.ToTVPair()], g.TasksToNodes[depTask])
				g.Graph.SetLine(line)

				g.EdgesToDeps[g.Graph.Edge(g.TasksToNodes[task.ToTVPair()].ID(), g.TasksToNodes[depTask].ID())] = dep
			}
		}
	}
}

// dependenciesForTaskUnit returns a slice of the task variant pairs this task unit depends on.
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

func (g *taskDependencyGraph) HasCycles() bool {
	return len(topo.TarjanSCC(g.Graph)) < g.Graph.Nodes().Len()
}

func (g *taskDependencyGraph) Cycles() [][]TVPair {
	var cycles [][]TVPair
	stronglyConnectedComponenets := topo.TarjanSCC(g.Graph)
	for _, scc := range stronglyConnectedComponenets {
		var cycle []TVPair
		for _, node := range scc {
			taskInCycle := g.NodesToTasks[node]
			cycle = append(cycle, taskInCycle)
		}
		cycles = append(cycles, cycle)
	}

	return cycles
}

func (g *taskDependencyGraph) AddEdge(dependentTask, dependedOnTask TVPair) {
	dependentNode := g.TasksToNodes[dependentTask]
	dependedOnNode := g.TasksToNodes[dependedOnTask]
	if dependentNode == nil || dependedOnNode == nil {
		return
	}

	line := g.Graph.NewLine(dependentNode, dependedOnNode)
	g.Graph.SetLine(line)
}

func (g *taskDependencyGraph) TasksDependingOnTask(t TVPair) []TVPair {
	dependedOnNode := g.TasksToNodes[t]
	if dependedOnNode == nil {
		return nil
	}

	var dependentTasks []TVPair
	nodes := g.Graph.To(dependedOnNode.ID())
	for nodes.Next() {
		dependentTasks = append(dependentTasks, g.NodesToTasks[nodes.Node()])
	}

	return dependentTasks
}

func (g *taskDependencyGraph) GetDependencyEdge(dependentTask, dependedOnTask TVPair) (TaskUnitDependency, error) {
	edge := g.Graph.Edge(g.TasksToNodes[dependentTask].ID(), g.TasksToNodes[dependedOnTask].ID())
	if edge == nil {
		return TaskUnitDependency{}, errors.Errorf("'%s' has no dependency on '%s'", dependentTask, dependedOnTask)
	}

	return g.EdgesToDeps[edge], nil
}
