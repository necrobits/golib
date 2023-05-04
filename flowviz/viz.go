package flowviz

import (
	"bytes"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
	"github.com/necrobits/golib/flow"
)

const (
	VizFormatDot graphviz.Format = "dot"
	VizFormatPNG graphviz.Format = graphviz.PNG
	VizFormatSVG graphviz.Format = graphviz.SVG
	VizFormatJPG graphviz.Format = graphviz.JPG
)

// CreateGraphvizForFlow creates a graphviz graph for the given flow.TransitionTable
// and writes it to the given buffer.
// Supported formats are: VizFormatDot, VizFormatPNG, VizFormatSVG, VizFormatJPG
func CreateGraphvizForFlow(transitionTable flow.TransitionTable, format graphviz.Format, buffer *bytes.Buffer) error {
	g := graphviz.New()
	graph, err := g.Graph()
	if err != nil {
		return err
	}
	for state, stateConfig := range transitionTable {
		for event, nextState := range stateConfig.Transitions {
			sName := string(state)
			tName := string(nextState)
			e := string(event)
			var sNode, tNode *cgraph.Node
			var edge *cgraph.Edge
			if sNode, err = graph.CreateNode(sName); err != nil {
				return err
			}
			if tNode, err = graph.CreateNode(tName); err != nil {
				return err
			}
			if edge, err = graph.CreateEdge(e, sNode, tNode); err != nil {
				return err
			}
			edge.SetLabel(e)
		}
	}
	if err := g.Render(graph, format, buffer); err != nil {
		return err
	}
	return nil
}
