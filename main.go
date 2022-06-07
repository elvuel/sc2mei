package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"golang.design/x/clipboard"
)

const (
	cols = 4
)

var (
	clipCtx             context.Context
	clipWatchCancelFunc context.CancelFunc

	list     []*ClipData
	listGrid *fyne.Container
	w        fyne.Window
)

func init() {
	list = make([]*ClipData, 0)
	listGrid = container.New(layout.NewGridLayout(cols), []fyne.CanvasObject{}...)
}

func main() {

	a := app.New()
	w = a.NewWindow("Screen capture to Markdown embed image")
	w.Resize(fyne.NewSize(400, 300))

	watchBtn := widget.NewButton("watch", func() {})
	watchBtn.OnTapped = func() {
		switch watchBtn.Text {
		case "watch":
			clipCtx, clipWatchCancelFunc = context.WithCancel(context.Background())
			go clipboardWatch()
			watchBtn.SetText("stop")
		case "stop":
			clipWatchCancelFunc()
			watchBtn.SetText("watch")
		}
	}

	w.SetContent(container.NewVBox(
		watchBtn,
		listGrid,
	))

	w.ShowAndRun()
}

type ClipData struct {
	T    time.Time
	Data []byte

	objs []fyne.CanvasObject
}

func (cd *ClipData) BuildObjects() {
	cd.objs = make([]fyne.CanvasObject, 0)
	cd.objs = append(cd.objs, widget.NewLabel(cd.T.Format("2006-01-02 15:04:05")))
	cd.objs = append(cd.objs, widget.NewButton("Show Image", func() {
		img := canvas.NewImageFromReader(bytes.NewBuffer(cd.Data), fmt.Sprintf("%d", cd.T.Unix()))
		img.FillMode = canvas.ImageFillOriginal
		dialog.ShowCustom(cd.T.Format("2006-01-02 15:04:05"), "close", img, w)
	}))
	cd.objs = append(cd.objs, widget.NewButton("Copy MD Base64", func() {
		buf := bytes.NewBufferString("![sc2mei]")
		buf.WriteString("(data:image/png;base64,")
		buf.WriteString(base64.StdEncoding.EncodeToString(cd.Data))
		buf.WriteString(")")
		clipboard.Write(clipboard.FmtText, buf.Bytes())
	}))
	cd.objs = append(cd.objs, widget.NewButton("Delete", func() {
		for i, v := range list {
			if v == cd {
				list = append(list[:i], list[i+1:]...)
				break
			}
		}
		cd.RemoveFrom(listGrid)
	}))
}

func (cd *ClipData) AddTo(c *fyne.Container) {
	for _, obj := range cd.objs {
		c.AddObject(obj)
	}
	c.Refresh()
}

func (cd *ClipData) RemoveFrom(c *fyne.Container) {
	for _, obj := range cd.objs {
		c.Remove(obj)
	}
	c.Refresh()
}

func clipboardWatch() {
	ch := clipboard.Watch(clipCtx, clipboard.FmtImage)
	for data := range ch {
		cd := &ClipData{
			T:    time.Now(),
			Data: data,
		}
		cd.BuildObjects()
		list = append(list, cd)
		cd.AddTo(listGrid)
	}
}
