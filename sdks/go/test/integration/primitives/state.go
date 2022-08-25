// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package primitives

import (
	"fmt"
	"strings"

	"github.com/apache/beam/sdks/v2/go/pkg/beam"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/core/state"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/register"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/testing/passert"
)

func init() {
	register.DoFn3x1[state.Provider, string, int, string](&valueStateFn{})
	register.DoFn3x1[state.Provider, string, int, string](&bagStateFn{})
	register.Emitter2[string, int]()
}

type valueStateFn struct {
	State1 state.Value[int]
	State2 state.Value[string]
}

func (f *valueStateFn) ProcessElement(s state.Provider, w string, c int) string {
	i, ok, err := f.State1.Read(s)
	if err != nil {
		panic(err)
	}
	if !ok {
		i = 1
	}
	err = f.State1.Write(s, i+1)
	if err != nil {
		panic(err)
	}

	j, ok, err := f.State2.Read(s)
	if err != nil {
		panic(err)
	}
	if !ok {
		j = "I"
	}
	err = f.State2.Write(s, j+"I")
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s: %v, %s", w, i, j)
}

// ValueStateParDo tests a DoFn that uses value state.
func ValueStateParDo() *beam.Pipeline {
	p, s := beam.NewPipelineWithRoot()

	in := beam.Create(s, "apple", "pear", "peach", "apple", "apple", "pear")
	keyed := beam.ParDo(s, func(w string, emit func(string, int)) {
		emit(w, 1)
	}, in)
	counts := beam.ParDo(s, &valueStateFn{State1: state.MakeValueState[int]("key1"), State2: state.MakeValueState[string]("key2")}, keyed)
	passert.Equals(s, counts, "apple: 1, I", "pear: 1, I", "peach: 1, I", "apple: 2, II", "apple: 3, III", "pear: 2, II")

	return p
}

type bagStateFn struct {
	State1 state.Bag[int]
	State2 state.Bag[string]
}

func (f *bagStateFn) ProcessElement(s state.Provider, w string, c int) string {
	i, ok, err := f.State1.Read(s)
	if err != nil {
		panic(err)
	}
	if !ok {
		i = []int{}
	}
	err = f.State1.Add(s, 1)
	if err != nil {
		panic(err)
	}

	j, ok, err := f.State2.Read(s)
	if err != nil {
		panic(err)
	}
	if !ok {
		j = []string{}
	}
	err = f.State2.Add(s, "I")
	if err != nil {
		panic(err)
	}
	sum := 0
	for _, val := range i {
		sum += val
	}
	return fmt.Sprintf("%s: %v, %s", w, sum, strings.Join(j, ","))
}

// ValueStateParDo tests a DoFn that uses value state.
func BagStateParDo() *beam.Pipeline {
	p, s := beam.NewPipelineWithRoot()

	in := beam.Create(s, "apple", "pear", "peach", "apple", "apple", "pear")
	keyed := beam.ParDo(s, func(w string, emit func(string, int)) {
		emit(w, 1)
	}, in)
	counts := beam.ParDo(s, &bagStateFn{State1: state.MakeBagState[int]("key1"), State2: state.MakeBagState[string]("key2")}, keyed)
	passert.Equals(s, counts, "apple: 0, ", "pear: 0, ", "peach: 0, ", "apple: 1, I", "apple: 2, I,I", "pear: 1, I")

	return p
}
