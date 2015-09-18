// Copyright 2015 Zack Guo <gizak@icloud.com>. All rights reserved.
// Use of this source code is governed by a MIT license that can
// be found in the LICENSE file.

package termui

import (
	"time"

	tm "github.com/nsf/termbox-go"
)

// Bufferer should be implemented by all renderable components.
type Bufferer interface {
	Buffer() Buffer
}

// Init initializes termui library. This function should be called before any others.
// After initialization, the library must be finalized by 'Close' function.
func Init() error {
	if err := tm.Init(); err != nil {
		return err
	}

	sysEvtChs = make([]chan Event, 0)
	go hookTermboxEvt()
	renderJobs = make(chan []Bufferer)
	go func() {
		for bs := range renderJobs {
			Render(bs...)
		}
	}()

	DefaultEvtStream.Init()
	DefaultEvtStream.Merge("termbox", NewSysEvtCh())
	DefaultEvtStream.Merge("timer", NewTimerCh(time.Second))
	DefaultEvtStream.Handle("/", DefualtHandler)

	return nil
}

// Close finalizes termui library,
// should be called after successful initialization when termui's functionality isn't required anymore.
func Close() {
	tm.Close()
}

// TermWidth returns the current terminal's width.
func TermWidth() int {
	tm.Sync()
	w, _ := tm.Size()
	return w
}

// TermHeight returns the current terminal's height.
func TermHeight() int {
	tm.Sync()
	_, h := tm.Size()
	return h
}

// Render renders all Bufferer in the given order from left to right,
// right could overlap on left ones.
func Render(bs ...Bufferer) {
	// set tm bg
	tm.Clear(tm.ColorDefault, toTmAttr(theme.BodyBg))
	for _, b := range bs {
		buf := b.Buffer()
		// set cels in buf
		for p, c := range buf.CellMap {
			if p.In(buf.Area) {
				tm.SetCell(p.X, p.Y, c.Ch, toTmAttr(c.Fg), toTmAttr(c.Bg))
			}
		}
	}
	// render
	tm.Flush()
}

var renderJobs chan []Bufferer

func SendBufferToRender(bs ...Bufferer) {
	go func() { renderJobs <- bs }()
}
