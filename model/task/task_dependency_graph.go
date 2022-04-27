package task

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/multi"
	"gonum.org/v1/gonum/graph/topo"
	"gonum.org/v1/gonum/graph/traverse"
)

type DependencyGraph struct {
	graph        *multi.DirectedGraph
	tasksToNodes map[TaskNode]graph.Node
	nodesToTasks map[graph.Node]TaskNode
	edgesToDeps  map[graph.Edge]DependencyEdge
}

type DependencyEdge struct {
	Status string
}

type TaskNode struct {
	Name    string
	Variant string

	ID string
}

func (t TaskNode) String() string {
	if t.ID != "" {
		return t.ID
	}

	return fmt.Sprintf("%s/%s", t.Variant, t.Name)
}

func VersionDependencyGraph(versionID string) (DependencyGraph, error) {
	g := NewDependencyGraph()
	tasks, err := FindAllTasksFromVersionWithDependencies(versionID)
	if err != nil {
		return g, errors.Wrapf(err, "getting tasks for version '%s'", versionID)
	}

	g.buildFromTasks(tasks)
	return g, nil
}

func NewDependencyGraph() DependencyGraph {
	return DependencyGraph{
		graph:        multi.NewDirectedGraph(),
		tasksToNodes: make(map[TaskNode]graph.Node),
		nodesToTasks: make(map[graph.Node]TaskNode),
		edgesToDeps:  make(map[graph.Edge]DependencyEdge),
	}
}

func (g *DependencyGraph) buildFromTasks(tasks []Task) error {
	taskIDToNode := make(map[string]TaskNode)
	for _, task := range tasks {
		tNode := task.ToTaskNode()
		g.AddTaskNode(tNode)
		taskIDToNode[task.Id] = tNode
	}

	for _, task := range tasks {
		dependentTaskNode := task.ToTaskNode()
		for _, dep := range task.DependsOn {
			dependedOnTaskNode := taskIDToNode[dep.TaskId]
			g.AddEdge(dependentTaskNode, dependedOnTaskNode, DependencyEdge{Status: dep.Status})
		}
	}

	return nil
}

func (g *DependencyGraph) AddTaskNode(tNode TaskNode) {
	node := g.graph.NewNode()
	g.graph.AddNode(node)
	g.tasksToNodes[tNode] = node
	g.nodesToTasks[node] = tNode
}

func (g *DependencyGraph) AddEdge(dependentTask, dependedOnTask TaskNode, dep DependencyEdge) {
	dependentNode := g.tasksToNodes[dependentTask]
	dependedOnNode := g.tasksToNodes[dependedOnTask]
	if dependentNode == nil || dependedOnNode == nil {
		return
	}

	line := g.graph.NewLine(dependentNode, dependedOnNode)
	g.graph.SetLine(line)
	g.edgesToDeps[g.graph.Edge(g.tasksToNodes[dependentTask].ID(), g.tasksToNodes[dependedOnTask].ID())] = dep
}

func (g *DependencyGraph) TasksDependingOnTask(t TaskNode) []TaskNode {
	dependedOnNode := g.tasksToNodes[t]
	if dependedOnNode == nil {
		return nil
	}

	var dependentTasks []TaskNode
	nodes := g.graph.To(dependedOnNode.ID())
	for nodes.Next() {
		dependentTasks = append(dependentTasks, g.nodesToTasks[nodes.Node()])
	}

	return dependentTasks
}

func (g *DependencyGraph) GetDependencyEdge(dependentTask, dependedOnTask TaskNode) (DependencyEdge, error) {
	edge := g.graph.Edge(g.tasksToNodes[dependentTask].ID(), g.tasksToNodes[dependedOnTask].ID())
	if edge == nil {
		return DependencyEdge{}, errors.Errorf("'%s' has no dependency on '%s'", dependentTask, dependedOnTask)
	}

	return g.edgesToDeps[edge], nil
}

type DependencyCycles [][]TaskNode

func (dc DependencyCycles) String() string {
	cycles := make([]string, 0, len(dc))
	for _, cycle := range dc {
		cycleStrings := make([]string, 0, len(cycle))
		for _, node := range cycle {
			cycleStrings = append(cycleStrings, node.String())
		}
		cycles = append(cycles, "[", strings.Join(cycleStrings, ", "), "]")
	}

	return strings.Join(cycles, ", ")
}

func (g *DependencyGraph) Cycles() DependencyCycles {
	var cycles DependencyCycles
	stronglyConnectedComponenets := topo.TarjanSCC(g.graph)
	for _, scc := range stronglyConnectedComponenets {
		if len(scc) <= 1 {
			continue
		}

		var cycle []TaskNode
		for _, node := range scc {
			taskInCycle := g.nodesToTasks[node]
			cycle = append(cycle, taskInCycle)
		}
		cycles = append(cycles, cycle)
	}

	return cycles
}

func (g *DependencyGraph) DepthFirstSearch(start, target TaskNode, traverseEdge func(dependentTask, dependedOnTask TaskNode, edge DependencyEdge) bool) bool {
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

func (g *DependencyGraph) TopologicalStableSort() ([]TaskNode, error) {
	sortedNodes, err := topo.SortStabilized(g.graph, nil)

	if err != nil {
		_, ok := err.(topo.Unorderable)
		if !ok {
			return nil, errors.Wrap(err, "sorting the graph")
		}
	}

	sortedTasks := make([]TaskNode, 0, len(sortedNodes))
	for _, node := range sortedNodes {
		if node != nil {
			sortedTasks = append(sortedTasks, g.nodesToTasks[node])
		}
	}

	return sortedTasks, nil
}
