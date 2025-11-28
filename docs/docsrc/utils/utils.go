package utils

import (
	"fmt"

	"github.com/shurcooL/sanitized_anchor_name"
	"github.com/sunfmin/snippetgo/parse"
	. "github.com/theplant/htmlgo"
	"github.com/theplant/osenv"
)

func Anchor(h *HTMLTagBuilder, text string) HTMLComponent {
	anchorName := sanitized_anchor_name.Create(text)
	return h.Children(
		Text(text),
		A().Class("anchor").Href(fmt.Sprintf("#%s", anchorName)),
	).Id(anchorName)
}

type Example struct {
	Title      string
	DemoPath   string
	SourcePath string
}

var LiveExamples []*Example

var envGitBranch = osenv.Get("GIT_BRANCH", "demo source code link git branch", "main")

func DemoWithSnippetLocation(title, demoPath string, location parse.Location) HTMLComponent {
	return Demo(title, demoPath, fmt.Sprintf("%s#L%d-L%d", location.File, location.StartLine, location.EndLine))
}

func Demo(title, demoPath, sourcePath string) HTMLComponent {
	if sourcePath != "" {
		sourcePath = fmt.Sprintf("https://github.com/qor5/admin/tree/%s/docs/docsrc/%s", envGitBranch, sourcePath)
	}
	ex := &Example{
		Title:      title,
		DemoPath:   demoPath,
		SourcePath: sourcePath,
	}

	if title != "" {
		LiveExamples = append(LiveExamples, ex)
	}

	return Div(
		Div(
			A().Text("Check the demo").Href(ex.DemoPath).Target("_blank"),
			Iff(ex.SourcePath != "", func() HTMLComponent {
				return Components(
					Text(" | "),
					A().Text("Source on GitHub").
						Href(ex.SourcePath).
						Target("_blank"),
				)
			}),
		).Class("demo"),
	)
}

func ExamplesDoc() HTMLComponent {
	u := Ul()
	for _, le := range LiveExamples {
		u.AppendChildren(
			Li(
				A().Href(le.DemoPath).Text(le.Title).Target("_blank"),
				Text(" | "),
				A().Href(le.SourcePath).Text("Source").Target("_blank"),
			),
		)
	}
	return u
}
