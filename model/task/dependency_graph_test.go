package task

import (
	"fmt"
	"testing"

	"github.com/evergreen-ci/evergreen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskNodeString(t *testing.T) {
	assert.Equal(t, "t0", TaskNode{ID: "t0", Name: "task0", Variant: "BV0"}.String())
	assert.Equal(t, "BV0/task0", TaskNode{Name: "task0", Variant: "BV0"}.String())
}

func TestBuildFromTasks(t *testing.T) {
	tasks := []Task{
		{
			Id: "t0",
			DependsOn: []Dependency{
				{
					TaskId: "t1",
					Status: evergreen.TaskSucceeded,
				},
			},
		},
		{
			Id: "t1",
			DependsOn: []Dependency{
				{TaskId: "t2"},
				{TaskId: "t3"},
			},
		},
		{Id: "t2"},
		{Id: "t3"},
	}

	t.Run("ForwardEdges", func(t *testing.T) {
		g := NewDependencyGraph()
		g.buildFromTasks(tasks, false)

		for _, task := range tasks {
			require.Contains(t, g.tasksToNodes, task.ToTaskNode())
			assert.Equal(t, task.ToTaskNode(), g.nodesToTasks[g.tasksToNodes[task.ToTaskNode()]])
		}

		assert.Equal(t, 0, g.graph.To(g.tasksToNodes[tasks[0].ToTaskNode()].ID()).Len())
		assert.Equal(t, 1, g.graph.From(g.tasksToNodes[tasks[0].ToTaskNode()].ID()).Len())

		assert.Equal(t, 1, g.graph.To(g.tasksToNodes[tasks[1].ToTaskNode()].ID()).Len())
		assert.Equal(t, 2, g.graph.From(g.tasksToNodes[tasks[1].ToTaskNode()].ID()).Len())

		assert.Equal(t, 1, g.graph.To(g.tasksToNodes[tasks[2].ToTaskNode()].ID()).Len())
		assert.Equal(t, 0, g.graph.From(g.tasksToNodes[tasks[2].ToTaskNode()].ID()).Len())

		assert.Equal(t, 1, g.graph.To(g.tasksToNodes[tasks[3].ToTaskNode()].ID()).Len())
		assert.Equal(t, 0, g.graph.From(g.tasksToNodes[tasks[3].ToTaskNode()].ID()).Len())

		t0_t1Edge := g.graph.Edge(g.tasksToNodes[tasks[0].ToTaskNode()].ID(), g.tasksToNodes[tasks[1].ToTaskNode()].ID())
		require.NotNil(t, t0_t1Edge)
		assert.Equal(t, DependencyEdge{Status: evergreen.TaskSucceeded}, g.edgesToDependencies[t0_t1Edge])

		t1_t2Edge := g.graph.Edge(g.tasksToNodes[tasks[1].ToTaskNode()].ID(), g.tasksToNodes[tasks[2].ToTaskNode()].ID())
		require.NotNil(t, t1_t2Edge)
		assert.Equal(t, DependencyEdge{}, g.edgesToDependencies[t1_t2Edge])

		t1_t3Edge := g.graph.Edge(g.tasksToNodes[tasks[1].ToTaskNode()].ID(), g.tasksToNodes[tasks[3].ToTaskNode()].ID())
		require.NotNil(t, t1_t3Edge)
		assert.Equal(t, DependencyEdge{}, g.edgesToDependencies[t1_t3Edge])
	})

	t.Run("ReversedEdges", func(t *testing.T) {
		g := NewDependencyGraph()
		g.buildFromTasks(tasks, true)

		for _, task := range tasks {
			require.Contains(t, g.tasksToNodes, task.ToTaskNode())
			assert.Equal(t, task.ToTaskNode(), g.nodesToTasks[g.tasksToNodes[task.ToTaskNode()]])
		}

		assert.Equal(t, 1, g.graph.To(g.tasksToNodes[tasks[0].ToTaskNode()].ID()).Len())
		assert.Equal(t, 0, g.graph.From(g.tasksToNodes[tasks[0].ToTaskNode()].ID()).Len())

		assert.Equal(t, 2, g.graph.To(g.tasksToNodes[tasks[1].ToTaskNode()].ID()).Len())
		assert.Equal(t, 1, g.graph.From(g.tasksToNodes[tasks[1].ToTaskNode()].ID()).Len())

		assert.Equal(t, 0, g.graph.To(g.tasksToNodes[tasks[2].ToTaskNode()].ID()).Len())
		assert.Equal(t, 1, g.graph.From(g.tasksToNodes[tasks[2].ToTaskNode()].ID()).Len())

		assert.Equal(t, 0, g.graph.To(g.tasksToNodes[tasks[3].ToTaskNode()].ID()).Len())
		assert.Equal(t, 1, g.graph.From(g.tasksToNodes[tasks[3].ToTaskNode()].ID()).Len())

		t1_t0Edge := g.graph.Edge(g.tasksToNodes[tasks[1].ToTaskNode()].ID(), g.tasksToNodes[tasks[0].ToTaskNode()].ID())
		require.NotNil(t, t1_t0Edge)
		assert.Equal(t, DependencyEdge{Status: "success"}, g.edgesToDependencies[t1_t0Edge])

		t2_t1Edge := g.graph.Edge(g.tasksToNodes[tasks[2].ToTaskNode()].ID(), g.tasksToNodes[tasks[1].ToTaskNode()].ID())
		require.NotNil(t, t2_t1Edge)
		assert.Equal(t, DependencyEdge{}, g.edgesToDependencies[t2_t1Edge])

		t3_t1Edge := g.graph.Edge(g.tasksToNodes[tasks[3].ToTaskNode()].ID(), g.tasksToNodes[tasks[1].ToTaskNode()].ID())
		require.NotNil(t, t3_t1Edge)
		assert.Equal(t, DependencyEdge{}, g.edgesToDependencies[t3_t1Edge])
	})
}

func TestAddTaskNode(t *testing.T) {
	tasks := []Task{
		{Id: "t0"},
	}

	t.Run("NewNode", func(t *testing.T) {
		g := NewDependencyGraph()
		g.buildFromTasks(tasks, false)
		assert.Len(t, g.nodesToTasks, 1)
		assert.Len(t, g.tasksToNodes, 1)
		assert.Equal(t, 1, g.graph.Nodes().Len())

		g.AddTaskNode(TaskNode{ID: "t1"})
		assert.Len(t, g.nodesToTasks, 2)
		assert.Len(t, g.tasksToNodes, 2)
		assert.Equal(t, 2, g.graph.Nodes().Len())
	})

	t.Run("PreexistingNode", func(t *testing.T) {
		g := NewDependencyGraph()
		g.buildFromTasks(tasks, false)
		assert.Len(t, g.nodesToTasks, 1)
		assert.Len(t, g.tasksToNodes, 1)
		assert.Equal(t, 1, g.graph.Nodes().Len())

		g.AddTaskNode(TaskNode{ID: "t0"})
		assert.Len(t, g.nodesToTasks, 1)
		assert.Len(t, g.tasksToNodes, 1)
		assert.Equal(t, 1, g.graph.Nodes().Len())
	})
}

func TestAddEdgeToGraph(t *testing.T) {
	tasks := []Task{
		{Id: "t0"},
		{Id: "t1", DependsOn: []Dependency{{TaskId: "t2"}}},
		{Id: "t2"},
	}

	t.Run("NewEdge", func(t *testing.T) {
		g := NewDependencyGraph()
		g.buildFromTasks(tasks, false)
		assert.Equal(t, 1, g.graph.Edges().Len())
		assert.Len(t, g.edgesToDependencies, 1)

		g.addEdgeToGraph(tasks[0].ToTaskNode(), tasks[1].ToTaskNode(), DependencyEdge{})
		assert.Equal(t, 2, g.graph.Edges().Len())
		assert.True(t, g.graph.HasEdgeFromTo(g.tasksToNodes[tasks[0].ToTaskNode()].ID(), g.tasksToNodes[tasks[1].ToTaskNode()].ID()))
		assert.Len(t, g.edgesToDependencies, 2)
	})

	t.Run("PreexistingEdge", func(t *testing.T) {
		g := NewDependencyGraph()
		g.buildFromTasks(tasks, false)
		assert.Equal(t, 1, g.graph.Edges().Len())
		assert.Len(t, g.edgesToDependencies, 1)

		g.addEdgeToGraph(tasks[1].ToTaskNode(), tasks[2].ToTaskNode(), DependencyEdge{})
		assert.Equal(t, 1, g.graph.Edges().Len())
		assert.Len(t, g.edgesToDependencies, 1)
	})

	t.Run("EdgeToMissingNode", func(t *testing.T) {
		g := NewDependencyGraph()
		g.buildFromTasks(tasks, false)
		assert.Equal(t, 1, g.graph.Edges().Len())
		assert.Len(t, g.edgesToDependencies, 1)

		g.addEdgeToGraph(tasks[0].ToTaskNode(), TaskNode{ID: "t3"}, DependencyEdge{})
		assert.Equal(t, 1, g.graph.Edges().Len())
		assert.Len(t, g.edgesToDependencies, 1)
	})

	t.Run("Cyclic", func(t *testing.T) {
		g := NewDependencyGraph()
		g.buildFromTasks(tasks, false)
		assert.Equal(t, 1, g.graph.Edges().Len())
		assert.Len(t, g.edgesToDependencies, 1)

		g.addEdgeToGraph(tasks[0].ToTaskNode(), tasks[1].ToTaskNode(), DependencyEdge{})
		g.addEdgeToGraph(tasks[1].ToTaskNode(), tasks[0].ToTaskNode(), DependencyEdge{})
		assert.Equal(t, 3, g.graph.Edges().Len())
		assert.True(t, g.graph.HasEdgeFromTo(g.tasksToNodes[tasks[0].ToTaskNode()].ID(), g.tasksToNodes[tasks[1].ToTaskNode()].ID()))
		assert.True(t, g.graph.HasEdgeFromTo(g.tasksToNodes[tasks[1].ToTaskNode()].ID(), g.tasksToNodes[tasks[0].ToTaskNode()].ID()))
		assert.Len(t, g.edgesToDependencies, 3)
	})
}

func TestGetDependencyEdge(t *testing.T) {
	tasks := []Task{
		{Id: "t0", DependsOn: []Dependency{{TaskId: "t1"}}},
		{Id: "t1", DependsOn: []Dependency{{TaskId: "t2", Status: evergreen.TaskSucceeded}}},
		{Id: "t2"},
	}
	g := NewDependencyGraph()
	g.buildFromTasks(tasks, false)

	t.Run("ExistingEdgeWithStatus", func(t *testing.T) {
		edge := g.GetDependencyEdge(tasks[1].ToTaskNode(), tasks[2].ToTaskNode())
		require.NotNil(t, edge)
		assert.Equal(t, evergreen.TaskSucceeded, edge.Status)
	})

	t.Run("ExistingEdgeNoStatus", func(t *testing.T) {
		edge := g.GetDependencyEdge(tasks[0].ToTaskNode(), tasks[1].ToTaskNode())
		require.NotNil(t, edge)
		assert.Empty(t, edge.Status)
	})

	t.Run("NonexistentEdge", func(t *testing.T) {
		edge := g.GetDependencyEdge(tasks[2].ToTaskNode(), tasks[0].ToTaskNode())
		assert.Nil(t, edge)
	})
}

func TestTasksDependingOnTask(t *testing.T) {
	tasks := []Task{
		{Id: "t0", DependsOn: []Dependency{{TaskId: "t1"}}},
		{Id: "t1"},
	}
	g := NewDependencyGraph()
	g.buildFromTasks(tasks, false)

	assert.Empty(t, g.TasksPointingToTask(tasks[0].ToTaskNode()))
	deps := g.TasksPointingToTask(tasks[1].ToTaskNode())
	require.Len(t, deps, 1)
	assert.Equal(t, tasks[0].Id, deps[0].ID)
}

func TestReachableFromNode(t *testing.T) {
	tasks := []Task{
		{Id: "t0", DependsOn: []Dependency{{TaskId: "t1"}, {TaskId: "t2"}}},
		{Id: "t1", DependsOn: []Dependency{{TaskId: "t3"}}},
		{Id: "t2", DependsOn: []Dependency{{TaskId: "t4"}}},
		{Id: "t3"},
		{Id: "t4"},
	}
	g := NewDependencyGraph()
	g.buildFromTasks(tasks, false)

	reachable := g.reachableFromNode(tasks[0].ToTaskNode())
	assert.Len(t, reachable, 4)
	assert.Contains(t, reachable, tasks[0].ToTaskNode())
	assert.Contains(t, reachable, tasks[1].ToTaskNode())
	assert.Contains(t, reachable, tasks[2].ToTaskNode())
	assert.Contains(t, reachable, tasks[3].ToTaskNode())

	reachable = g.reachableFromNode(tasks[1].ToTaskNode())
	assert.Len(t, reachable, 2)
	assert.Contains(t, reachable, tasks[1].ToTaskNode())
	assert.Contains(t, reachable, tasks[3].ToTaskNode())

	reachable = g.reachableFromNode(tasks[3].ToTaskNode())
	assert.Empty(t, reachable)
}

func TestCycles(t *testing.T) {
	t.Run("EmptyGraph", func(t *testing.T) {
		g := NewDependencyGraph()
		assert.Empty(t, g.Cycles())
	})

	t.Run("NoCycles", func(t *testing.T) {
		tasks := []Task{
			{Id: "t0", DependsOn: []Dependency{{TaskId: "t1"}}},
			{Id: "t1"},
		}
		g := NewDependencyGraph()
		g.buildFromTasks(tasks, false)

		assert.Empty(t, g.Cycles())
	})

	t.Run("Loops", func(t *testing.T) {
		tasks := []Task{
			{Id: "t0", DependsOn: []Dependency{{TaskId: "t0"}}},
		}
		g := NewDependencyGraph()
		g.buildFromTasks(tasks, false)

		assert.Empty(t, g.Cycles())
	})

	t.Run("TwoConnectedCycles", func(t *testing.T) {
		tasks := []Task{
			{Id: "t0", DependsOn: []Dependency{{TaskId: "t1"}}},
			{Id: "t1", DependsOn: []Dependency{{TaskId: "t0"}, {TaskId: "t2"}}},
			{Id: "t2", DependsOn: []Dependency{{TaskId: "t3"}}},
			{Id: "t3", DependsOn: []Dependency{{TaskId: "t2"}}},
		}
		g := NewDependencyGraph()
		g.buildFromTasks(tasks, false)

		cycles := g.Cycles()
		assert.Len(t, cycles, 2)
	})

	t.Run("TwoDisconnectedCycles", func(t *testing.T) {
		tasks := []Task{
			{Id: "t0", DependsOn: []Dependency{{TaskId: "t1"}}},
			{Id: "t1", DependsOn: []Dependency{{TaskId: "t0"}}},
			{Id: "t2", DependsOn: []Dependency{{TaskId: "t3"}}},
			{Id: "t3", DependsOn: []Dependency{{TaskId: "t2"}}},
		}
		g := NewDependencyGraph()
		g.buildFromTasks(tasks, false)

		cycles := g.Cycles()
		assert.Len(t, cycles, 2)
	})
}

func TestDependencyCyclesString(t *testing.T) {
	t.Run("NoCycles", func(t *testing.T) {
		dc := DependencyCycles{}
		assert.Empty(t, dc.String())
	})

	t.Run("OneCycle", func(t *testing.T) {
		ids := []string{"t0", "t1"}
		dc := DependencyCycles{
			{{ID: ids[0]}, {ID: ids[1]}},
		}
		assert.Equal(t, fmt.Sprintf("[%s, %s]", ids[0], ids[1]), dc.String())
	})

	t.Run("TwoCycles", func(t *testing.T) {
		ids := []string{"t0", "t1", "t3", "t4"}
		dc := DependencyCycles{
			{{ID: ids[0]}, {ID: ids[1]}},
			{{ID: ids[3]}, {ID: ids[4]}},
		}
		assert.Equal(t, fmt.Sprintf("[%s, %s], [%s, %s]", ids[0], ids[1], ids[3], ids[4]), dc.String())
	})
}

func TestDepthFirstSearch(t *testing.T) {
	tasks := []Task{
		{Id: "t0", DependsOn: []Dependency{{TaskId: "t1", Status: evergreen.TaskSucceeded}}},
		{Id: "t1", DependsOn: []Dependency{{TaskId: "t2"}}},
		{Id: "t2"},
		{Id: "t3"},
	}
	g := NewDependencyGraph()
	g.buildFromTasks(tasks, false)

	t.Run("NilTraverseEdge", func(t *testing.T) {
		assert.True(t, g.DepthFirstSearch(tasks[0].ToTaskNode(), tasks[2].ToTaskNode(), nil))
		assert.False(t, g.DepthFirstSearch(tasks[1].ToTaskNode(), tasks[0].ToTaskNode(), nil))
		assert.False(t, g.DepthFirstSearch(tasks[3].ToTaskNode(), tasks[0].ToTaskNode(), nil))
	})

	t.Run("TraversalBlockedAtNode", func(t *testing.T) {
		assert.False(t, g.DepthFirstSearch(tasks[0].ToTaskNode(), tasks[2].ToTaskNode(), func(currentNode, nextNode TaskNode, edge DependencyEdge) bool {
			if nextNode == tasks[1].ToTaskNode() {
				return false
			}
			return true
		}))
	})

	t.Run("TraversalBlockedAtEdge", func(t *testing.T) {
		assert.False(t, g.DepthFirstSearch(tasks[0].ToTaskNode(), tasks[2].ToTaskNode(), func(currentNode, nextNode TaskNode, edge DependencyEdge) bool {
			if edge.Status == evergreen.TaskSucceeded {
				return true
			}
			return false
		}))
	})

	t.Run("StartMissingFromGraph", func(t *testing.T) {
		assert.False(t, g.DepthFirstSearch(TaskNode{ID: "t4"}, tasks[0].ToTaskNode(), nil))
	})

	t.Run("TargetMissingFromGraph", func(t *testing.T) {
		assert.False(t, g.DepthFirstSearch(tasks[0].ToTaskNode(), TaskNode{ID: "t4"}, nil))
	})
}

func TestTopologicalStableSort(t *testing.T) {
	t.Run("StableSortNoDependencies", func(t *testing.T) {
		tasks := []Task{
			{Id: "t0"},
			{Id: "t1"},
			{Id: "t2"},
		}
		g := NewDependencyGraph()
		g.buildFromTasks(tasks, true)

		sortedNodes, err := g.TopologicalStableSort()
		assert.NoError(t, err)
		require.Len(t, sortedNodes, 3)
		assert.Equal(t, tasks[0].ToTaskNode(), sortedNodes[0])
		assert.Equal(t, tasks[1].ToTaskNode(), sortedNodes[1])
		assert.Equal(t, tasks[2].ToTaskNode(), sortedNodes[2])
	})

	t.Run("StableSortWithDependencies", func(t *testing.T) {
		tasks := []Task{
			{Id: "t0"},
			{Id: "t1", DependsOn: []Dependency{{TaskId: "t0"}}},
			{Id: "t2"},
		}
		g := NewDependencyGraph()
		g.buildFromTasks(tasks, true)

		sortedNodes, err := g.TopologicalStableSort()
		assert.NoError(t, err)
		require.Len(t, sortedNodes, 3)
		assert.Equal(t, tasks[1].ToTaskNode(), sortedNodes[0])
		assert.Equal(t, tasks[0].ToTaskNode(), sortedNodes[1])
		assert.Equal(t, tasks[2].ToTaskNode(), sortedNodes[2])
	})

	t.Run("Cycle", func(t *testing.T) {
		tasks := []Task{
			{Id: "t0", DependsOn: []Dependency{{TaskId: "t1"}}},
			{Id: "t1", DependsOn: []Dependency{{TaskId: "t0"}}},
			{Id: "t2"},
		}
		g := NewDependencyGraph()
		g.buildFromTasks(tasks, true)

		sortedNodes, err := g.TopologicalStableSort()
		assert.NoError(t, err)
		require.Len(t, sortedNodes, 1)
		assert.Equal(t, tasks[2].ToTaskNode(), sortedNodes[0])
	})

	t.Run("EmptyGraph", func(t *testing.T) {
		g := NewDependencyGraph()

		sortedNodes, err := g.TopologicalStableSort()
		assert.NoError(t, err)
		assert.Empty(t, sortedNodes)
	})
}
