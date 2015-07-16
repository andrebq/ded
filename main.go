package main

import (
	"amoraes.info/ded/vfs"
	"amoraes.info/ded/vfs/namespace"
	"bytes"
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/andrebq/gas"
	"github.com/google/gxui"
	"github.com/google/gxui/drivers/gl"
	"github.com/google/gxui/gxfont"
	"github.com/google/gxui/math"
	"github.com/google/gxui/themes/dark"
	_ "time"
)

type (
	MyTheme struct {
		gxui.Theme

		monospacedFont gxui.Font

		initialized bool
	}
)

func CreateMyTheme(parent gxui.Theme) *MyTheme {
	mt := &MyTheme{
		Theme: parent,
	}
	mt.Init()
	return mt
}

func (mt *MyTheme) Init() {
	if mt.initialized {
		return
	}

	defaultMonospaceFont, err := mt.Driver().CreateFont(gxfont.Monospace, 12)
	if err == nil {
		defaultMonospaceFont.LoadGlyphs(32, 126)
	} else {
		panic(fmt.Sprintf("Warning: Failed to load default monospace font - %v", err))
	}
	mt.monospacedFont = defaultMonospaceFont

	mt.initialized = true
}

func (mt *MyTheme) CreateDedEditor(singleline bool) *DedEditor {
	ml := &DedEditor{font: mt.monospacedFont, oneLine: singleline}
	ml.Init(ml, mt)
	return ml
}

func (mt *MyTheme) CreateEditorLayout() *EditorLayout {
	el := &EditorLayout{}
	el.Init(el, mt)
	return el
}

func appMain(driver gxui.Driver) {
	theme := CreateMyTheme(dark.CreateTheme(driver))

	window := theme.CreateWindow(800, 600, "Hi")
	//window.SetBackgroundBrush(gxui.CreateBrush(gxui.Gray50))

	dedEditor := theme.CreateDedEditor(false)
	dedEditor.SetFont(nil, 14)

	func() {
		txt, err := gas.ReadFile("amoraes.info/ded/deditor.go")
		if err != nil {
			panic(err)
		}
		dedEditor.SetText(string(txt))
	}()
	dedEditor.SetText("just for a test")

	dedEditorBar := theme.CreateDedEditor(true)
	dedEditorBar.SetFont(nil, 16)
	dedEditorBar.SetText("Save | Reload | Quit")

	el := theme.CreateEditorLayout()
	el.AddChild(dedEditorBar)
	el.AddChild(dedEditor)
	el.SetChildSize(dedEditorBar, math.Size{H: dedEditorBar.LineHeight()})

	editorfs := &EditorFS{
		editor: dedEditor,
		bar:    dedEditorBar,
	}
	log.Infof("Exporting fs")
	err := editorfs.ExportAt(&dedNamespace, "active")
	log.Infof("EditorFS exported")
	if err != nil {
		log.Fatalf("Unable to export editor fs: %v", err)
	}

	/*
		TODO(andre): this code requires the proper namespace implementation,
		don't expose the FS.
		editorfs := NewEditorFS(dedEditor, dedEditorBar)
		srv, err := vfs.NewTCPServer(&vfs.Fileserver{editorfs}, "0.0.0.0:5640")
		if err != nil {
			log.Printf("Error starting server: %v", err)
			window.Close()
			return
		}
		window.OnClose(func() { srv.Close() })
	*/

	ll := theme.CreateLinearLayout()
	ll.SetDirection(gxui.TopToBottom)
	ll.AddChild(el)

	window.AddChild(ll)
	gxui.SetFocus(dedEditor)

	window.OnClose(driver.Terminate)
}

var (
	rootCli *vfs.Client

	dedNamespace namespace.Namespace
	listenAddr   = flag.String("addr", ":5640", "Address to listen for incoming data")
)

func main() {
	log.SetLevel(log.DebugLevel)
	fmt.Printf("")

	export := namespace.NewExport(&dedNamespace)
	srv, err := vfs.NewTCPServer(&vfs.Fileserver{export}, *listenAddr)
	if err != nil {
		log.Fatalf("Unable to start Ded tcp server. %v", err)
	}
	_ = srv
	gl.StartDriver(appMain)
	srv.Close()
}

func println(vals ...interface{}) {
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "!PRINTLN!\t")
	for idx, v := range vals {
		if idx == 0 {
			fmt.Fprintf(buf, "%v", v)
			continue
		}
		fmt.Fprintf(buf, " %v", v)
	}

	log.Debugf("%v", string(buf.Bytes()))
}
