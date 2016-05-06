// Mgmt
// Copyright (C) 2013-2016+ James Shubin and the project contributors
// Written by James Shubin <james@shubin.ca> and the project contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import

//"github.com/go-fsnotify/fsnotify" // git master of "gopkg.in/fsnotify.v1"
(
	"encoding/gob"
	"log"
)

func init() {
	gob.Register(&TimerRes{})
}

type TimerRes struct {
	BaseRes `yaml:",inline"`
}

func NewTimerRes(name string) *TimerRes {
	obj := &TimerRes{
		BaseRes: BaseRes{
			Name: name,
		},
	}
	obj.Init()
	return obj
}

func (obj *TimerRes) Init() {
	obj.BaseRes.kind = "Timer"
	obj.BaseRes.Init() // call base init, b/c we're overriding
}

// validate if the params passed in are valid data
// FIXME: where should this get called ?
func (obj *TimerRes) Validate() bool {
	return true
}

func (obj *TimerRes) Watch(processChan chan Event) {
	if obj.IsWatching() {
		return
	}
	obj.SetWatching(true)
	defer obj.SetWatching(false)
	cuuid := obj.converger.Register()
	defer cuuid.Unregister()

	var send = false // send event?
	var exit = false
	for {
		obj.SetState(resStateWatching) // reset
		select {
		case event := <-obj.events:
			cuuid.SetConverged(false)
			// we avoid sending events on unpause
			if exit, send = obj.ReadEvent(&event); exit {
				return // exit
			}

		case _ = <-cuuid.ConvergedTimer():
			cuuid.SetConverged(true) // converged!
			continue
		}

		// do all our event sending all together to avoid duplicate msgs
		if send {
			send = false
			// only do this on certain types of events
			//obj.isStateOK = false // something made state dirty
			resp := NewResp()
			processChan <- Event{eventNil, resp, "", true} // trigger process
			resp.ACKWait()                                 // wait for the ACK()
		}
	}
}

// CheckApply method for Noop resource. Does nothing, returns happy!
func (obj *TimerRes) CheckApply(apply bool) (stateok bool, err error) {
	log.Printf("%v[%v]: CheckApply(%t)", obj.Kind(), obj.GetName(), apply)
	return true, nil // state is always okay
}

type TimerUUID struct {
	BaseUUID
	name string
}

func (obj *TimerRes) AutoEdges() AutoEdge {
	return nil
}

// include all params to make a unique identification of this object
// most resources only return one, although some resources return multiple
func (obj *TimerRes) GetUUIDs() []ResUUID {
	x := &TimerUUID{
		BaseUUID: BaseUUID{name: obj.GetName(), kind: obj.Kind()},
		name:     obj.Name,
	}
	return []ResUUID{x}
}

func (obj *TimerRes) GroupCmp(r Res) bool {
	_, ok := r.(*TimerRes)
	if !ok {
		return false
	}
	return true // timer resources can always be grouped together! PurpleIdea: Is this true for timer?
}

func (obj *TimerRes) Compare(res Res) bool {
	switch res.(type) {
	// we can only compare TimerRes to others of the same resource
	case *TimerRes:
		res := res.(*TimerRes)
		if obj.Name != res.Name {
			return false
		}
	default:
		return false
	}
	return true
}
