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
	tasksToNodes map[taskNode]graph.Node
	nodesToTasks map[graph.Node]taskNode
	edgesToDeps  map[graph.Edge]dependencyEdge
}

type dependencyEdge struct {
	status string
}

type taskNode struct {
	name    string
	variant string
	version string
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
		tasksToNodes: make(map[taskNode]graph.Node),
		nodesToTasks: make(map[graph.Node]taskNode),
		edgesToDeps:  make(map[graph.Edge]dependencyEdge),
	}
}

func (g *taskDependencyGraph) buildFromTasks(tasks []task.Task) error {
	taskIDToNode := make(map[string]taskNode)
	for _, task := range tasks {
		tNode := taskNode{
			name:    task.DisplayName,
			variant: task.BuildVariant,
			version: task.Version,
		}
		g.addTaskNode(tNode)
		taskIDToNode[task.Id] = tNode
	}

	for _, task := range tasks {
		dependentTaskNode := taskNode{name: task.DisplayName, variant: task.BuildVariant}
		for _, dep := range task.DependsOn {
			dependedOnTaskNode := taskIDToNode[dep.TaskId]
			g.addEdge(dependentTaskNode, dependedOnTaskNode, dependencyEdge{status: dep.Status})
		}
	}

	return nil
}

func (g *taskDependencyGraph) buildFromProject(p *Project) {
	tasks := p.FindAllBuildVariantTasks()
	var taskTVs []TVPair

	for _, task := range tasks {
		g.addTaskNode(taskNode{name: task.Name, variant: task.Variant})
		taskTVs = append(taskTVs, task.ToTVPair())
	}

	for _, task := range tasks {
		tNode := taskNode{name: task.Name, variant: task.Variant}
		for dep, depTasks := range dependenciesForTaskUnit(task, taskTVs) {
			for _, depNode := range depTasks {
				g.addEdge(tNode, depNode, dep)
			}
		}
	}
}

// dependenciesForTaskUnit returns a map of dependencies to the task variant pairs they match.
func dependenciesForTaskUnit(dependentTaskUnit BuildVariantTaskUnit, allTVPairs []TVPair) map[dependencyEdge][]taskNode {
	dependencies := make(map[dependencyEdge][]taskNode)
	for _, dep := range dependentTaskUnit.DependsOn {
		// Use the current variant if none is specified.
		if dep.Variant == "" {
			dep.Variant = dependentTaskUnit.Variant
		}

		for _, tv := range allTVPairs {
			depNode := taskNode{variant: tv.Variant, name: tv.TaskName}
			if tv != dependentTaskUnit.ToTVPair() &&
				(dep.Variant == AllVariants || depNode.variant == dep.Variant) &&
				(dep.Name == AllDependencies || depNode.name == dep.Name) {
				dependencies[dependencyEdge{status: dep.Status}] = append(dependencies[dependencyEdge{status: dep.Status}], depNode)
			}
		}
	}

	return dependencies
}

func (g *taskDependencyGraph) addTaskNode(tNode taskNode) {
	node := g.graph.NewNode()
	g.graph.AddNode(node)
	g.tasksToNodes[tNode] = node
	g.nodesToTasks[node] = tNode
}

func (g *taskDependencyGraph) addEdge(dependentTask, dependedOnTask taskNode, dep dependencyEdge) {
	dependentNode := g.tasksToNodes[dependentTask]
	dependedOnNode := g.tasksToNodes[dependedOnTask]
	if dependentNode == nil || dependedOnNode == nil {
		return
	}

	line := g.graph.NewLine(dependentNode, dependedOnNode)
	g.graph.SetLine(line)
	g.edgesToDeps[g.graph.Edge(g.tasksToNodes[dependentTask].ID(), g.tasksToNodes[dependedOnTask].ID())] = dep
}

func (g *taskDependencyGraph) tasksDependingOnTask(t taskNode) []taskNode {
	dependedOnNode := g.tasksToNodes[t]
	if dependedOnNode == nil {
		return nil
	}

	var dependentTasks []taskNode
	nodes := g.graph.To(dependedOnNode.ID())
	for nodes.Next() {
		dependentTasks = append(dependentTasks, g.nodesToTasks[nodes.Node()])
	}

	return dependentTasks
}

func (g *taskDependencyGraph) getDependencyEdge(dependentTask, dependedOnTask taskNode) (dependencyEdge, error) {
	edge := g.graph.Edge(g.tasksToNodes[dependentTask].ID(), g.tasksToNodes[dependedOnTask].ID())
	if edge == nil {
		return dependencyEdge{}, errors.Errorf("'%s' has no dependency on '%s'", dependentTask, dependedOnTask)
	}

	return g.edgesToDeps[edge], nil
}

func (g *taskDependencyGraph) Cycles() [][]taskNode {
	var cycles [][]taskNode
	stronglyConnectedComponenets := topo.TarjanSCC(g.graph)
	for _, scc := range stronglyConnectedComponenets {
		if len(scc) <= 1 {
			continue
		}

		var cycle []taskNode
		for _, node := range scc {
			taskInCycle := g.nodesToTasks[node]
			cycle = append(cycle, taskInCycle)
		}
		cycles = append(cycles, cycle)
	}

	return cycles
}

func (g *taskDependencyGraph) DepthFirstSearch(start, target taskNode, traverseEdge func(dependentTask, dependedOnTask taskNode, edge dependencyEdge) bool) bool {
	traversal := traverse.DepthFirst{
		Traverse: func(e graph.Edge) bool {
			if traverseEdge == nil {
				return true
			}

			dependedOn := g.nodesToTasks[e.From()]
			dependent := g.nodesToTasks[e.To()]
			edge := g.edgesToDeps[e]

			return traverseEdge(dependent, dependedOn, edge)
		},
	}

	return traversal.Walk(g.graph, g.tasksToNodes[start], func(n graph.Node) bool { return g.nodesToTasks[n] == target }) == nil
}
