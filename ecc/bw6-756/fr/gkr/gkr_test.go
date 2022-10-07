// Copyright 2020 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package gkr

import (
	"encoding/json"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bw6-756/fr"
	"github.com/consensys/gnark-crypto/ecc/bw6-756/fr/polynomial"
	"github.com/consensys/gnark-crypto/ecc/bw6-756/fr/sumcheck"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestNoGateTwoInstances(t *testing.T) {
	// Testing a single instance is not possible because the sumcheck implementation doesn't cover the trivial 0-variate case
	testNoGate(t, []fr.Element{four, three})
}

func TestNoGate(t *testing.T) {
	testManyInstances(t, 1, testNoGate)
}

func TestSingleMulGateTwoInstances(t *testing.T) {
	testSingleMulGate(t, []fr.Element{four, three}, []fr.Element{two, three})
}

func TestSingleMulGate(t *testing.T) {
	testManyInstances(t, 2, testSingleMulGate)
}

func TestSingleInputTwoIdentityGatesTwoInstances(t *testing.T) {

	testSingleInputTwoIdentityGates(t, []fr.Element{two, three})
}

func TestSingleInputTwoIdentityGates(t *testing.T) {

	testManyInstances(t, 2, testSingleInputTwoIdentityGates)
}

func TestSingleInputTwoIdentityGatesComposedTwoInstances(t *testing.T) {
	testSingleInputTwoIdentityGatesComposed(t, []fr.Element{two, one})
}

func TestSingleInputTwoIdentityGatesComposed(t *testing.T) {
	testManyInstances(t, 1, testSingleInputTwoIdentityGatesComposed)
}

func TestSingleMimcCipherGateTwoInstances(t *testing.T) {
	testSingleMimcCipherGate(t, []fr.Element{one, one}, []fr.Element{one, two})
}

func TestSingleMimcCipherGate(t *testing.T) {
	testManyInstances(t, 2, testSingleMimcCipherGate)
}

func TestATimesBSquaredTwoInstances(t *testing.T) {
	testATimesBSquared(t, 2, []fr.Element{one, one}, []fr.Element{one, two})
}

func TestShallowMimcTwoInstances(t *testing.T) {
	testMimc(t, 2, []fr.Element{one, one}, []fr.Element{one, two})
}

func TestMimcTwoInstances(t *testing.T) {
	testMimc(t, 93, []fr.Element{one, one}, []fr.Element{one, two})
}

func TestMimc(t *testing.T) {
	testManyInstances(t, 2, generateTestMimc(93))
}

func TestRecreateSumcheckErrorFromSingleInputTwoIdentityGatesGateTwoInstances(t *testing.T) {
	circuit := Circuit{{Wire{
		Gate:       nil,
		Inputs:     []*Wire{},
		NumOutputs: 2,
	}}}

	wire := &circuit[0][0]

	assignment := WireAssignment{&circuit[0][0]: []fr.Element{two, three}}

	claimsManagerGen := func() *claimsManager {
		manager := newClaimsManager(circuit, assignment)
		manager.add(wire, []fr.Element{three}, five)
		manager.add(wire, []fr.Element{four}, six)
		return &manager
	}

	transcriptGen := sumcheck.NewMessageCounterGenerator(4, 1)

	proof := sumcheck.Prove(claimsManagerGen().getClaim(wire), transcriptGen())
	sumcheck.Verify(claimsManagerGen().getLazyClaim(wire), proof, transcriptGen())
}

// complete the circuit evaluation from input values
func (a WireAssignment) complete(c Circuit) WireAssignment {
	numEvaluations := len(a[&c[len(c)-1][0]])

	for i := len(c) - 2; i >= 0; i-- { //there can only be input wires in the bottommost layer
		layer := c[i]
		for j := 0; j < len(layer); j++ {
			wire := &layer[j]

			if !wire.IsInput() {
				evals := make([]fr.Element, numEvaluations)
				ins := make([]fr.Element, len(wire.Inputs))
				for k := 0; k < numEvaluations; k++ {
					for inI, in := range wire.Inputs {
						ins[inI] = a[in][k]
					}
					evals[k] = wire.Gate.Evaluate(ins...)
				}
				a[wire] = evals
			}
		}
	}
	return a
}

var one, two, three, four, five, six fr.Element

func init() {
	one.SetOne()
	two.Double(&one)
	three.Add(&two, &one)
	four.Double(&two)
	five.Add(&three, &two)
	six.Double(&three)
}

var testManyInstancesLogMaxInstances = -1

func getLogMaxInstances(t *testing.T) int {
	if testManyInstancesLogMaxInstances == -1 {

		s := os.Getenv("GKR_LOG_INSTANCES")
		if s == "" {
			testManyInstancesLogMaxInstances = 5
		} else {
			var err error
			testManyInstancesLogMaxInstances, err = strconv.Atoi(s)
			if err != nil {
				t.Error(err)
			}
		}

	}
	return testManyInstancesLogMaxInstances
}

func testManyInstances(t *testing.T, numInput int, test func(*testing.T, ...[]fr.Element)) {
	fullAssignments := make([][]fr.Element, numInput)
	maxSize := 1 << getLogMaxInstances(t)

	t.Log("Entered test orchestrator, assigning and randomizing inputs")

	for i := range fullAssignments {
		fullAssignments[i] = polynomial.Make(maxSize)
		setRandom(fullAssignments[i])
	}

	defer polynomial.Dump(fullAssignments...)

	inputAssignments := make([][]fr.Element, numInput)
	for numEvals := maxSize; numEvals <= maxSize; numEvals *= 2 {
		for i, fullAssignment := range fullAssignments {
			inputAssignments[i] = fullAssignment[:numEvals]
		}

		t.Log("Selected inputs for test")
		test(t, inputAssignments...)
	}
}

func testNoGate(t *testing.T, inputAssignments ...[]fr.Element) {
	c := Circuit{
		{
			{
				Inputs:     []*Wire{},
				NumOutputs: 1,
				Gate:       nil,
			},
		},
	}

	assignment := WireAssignment{&c[0][0]: inputAssignments[0]}

	proof := Prove(c, assignment, sumcheck.NewMessageCounter(1, 1))

	// Even though a hash is called here, the proof is empty

	if !Verify(c, assignment, proof, sumcheck.NewMessageCounter(1, 1)) {
		t.Error("Proof rejected")
	}
}

func testSingleMulGate(t *testing.T, inputAssignments ...[]fr.Element) {
	c := make(Circuit, 2)

	c[1] = CircuitLayer{
		{
			Inputs:     []*Wire{},
			NumOutputs: 1,
			Gate:       nil,
		},
		{
			Inputs:     []*Wire{},
			NumOutputs: 1,
			Gate:       nil,
		},
	}

	c[0] = CircuitLayer{{
		Inputs:     []*Wire{&c[1][0], &c[1][1]},
		NumOutputs: 1,
		Gate:       mulGate{},
	}}

	assignment := WireAssignment{&c[1][0]: inputAssignments[0], &c[1][1]: inputAssignments[1]}.complete(c)

	proof := Prove(c, assignment, sumcheck.NewMessageCounter(1, 1))

	if !Verify(c, assignment, proof, sumcheck.NewMessageCounter(1, 1)) {
		t.Error("Proof rejected")
	}

	if Verify(c, assignment, proof, sumcheck.NewMessageCounter(0, 1)) {
		t.Error("Bad proof accepted")
	}
}

func testSingleInputTwoIdentityGates(t *testing.T, inputAssignments ...[]fr.Element) {
	c := make(Circuit, 2)

	c[1] = CircuitLayer{
		{
			Inputs:     []*Wire{},
			NumOutputs: 2,
			Gate:       nil,
		},
	}

	c[0] = CircuitLayer{
		{
			Inputs:     []*Wire{&c[1][0]},
			NumOutputs: 1,
			Gate:       identityGate{},
		},
		{
			Inputs:     []*Wire{&c[1][0]},
			NumOutputs: 1,
			Gate:       identityGate{},
		},
	}

	assignment := WireAssignment{&c[1][0]: inputAssignments[0]}.complete(c)

	proof := Prove(c, assignment, sumcheck.NewMessageCounter(0, 1))

	if !Verify(c, assignment, proof, sumcheck.NewMessageCounter(0, 1)) {
		t.Error("Proof rejected")
	}

	if Verify(c, assignment, proof, sumcheck.NewMessageCounter(1, 1)) {
		t.Error("Bad proof accepted")
	}
}

func testSingleMimcCipherGate(t *testing.T, inputAssignments ...[]fr.Element) {
	c := make(Circuit, 2)

	c[1] = CircuitLayer{
		{
			Inputs:     []*Wire{},
			NumOutputs: 1,
			Gate:       nil,
		},
		{
			Inputs:     []*Wire{},
			NumOutputs: 1,
			Gate:       nil,
		},
	}

	c[0] = CircuitLayer{
		{
			Inputs:     []*Wire{&c[1][0], &c[1][1]},
			NumOutputs: 1,
			Gate:       mimcCipherGate{},
		},
	}
	t.Log("Evaluating all circuit wires")
	assignment := WireAssignment{&c[1][0]: inputAssignments[0], &c[1][1]: inputAssignments[1]}.complete(c)
	t.Log("Circuit evaluation complete")
	proof := Prove(c, assignment, sumcheck.NewMessageCounter(0, 1))
	t.Log("Proof complete")
	if !Verify(c, assignment, proof, sumcheck.NewMessageCounter(0, 1)) {
		t.Error("Proof rejected")
	}
	t.Log("Successful verification complete")
	if Verify(c, assignment, proof, sumcheck.NewMessageCounter(1, 1)) {
		t.Error("Bad proof accepted")
	}
	t.Log("Unsuccessful verification complete")
}

func testSingleInputTwoIdentityGatesComposed(t *testing.T, inputAssignments ...[]fr.Element) {
	c := make(Circuit, 3)

	c[2] = CircuitLayer{{
		Gate:       nil,
		Inputs:     []*Wire{},
		NumOutputs: 1,
	}}
	c[1] = CircuitLayer{{
		Gate:       identityGate{},
		Inputs:     []*Wire{&c[2][0]},
		NumOutputs: 1,
	}}
	c[0] = CircuitLayer{{
		Gate:       identityGate{},
		Inputs:     []*Wire{&c[1][0]},
		NumOutputs: 1,
	}}

	assignment := WireAssignment{&c[2][0]: inputAssignments[0]}.complete(c)

	proof := Prove(c, assignment, sumcheck.NewMessageCounter(0, 1))

	if !Verify(c, assignment, proof, sumcheck.NewMessageCounter(0, 1)) {
		t.Error("Proof rejected")
	}

	if Verify(c, assignment, proof, sumcheck.NewMessageCounter(1, 1)) {
		t.Error("Bad proof accepted")
	}
}

func generateTestMimc(numRounds int) func(*testing.T, ...[]fr.Element) {
	return func(t *testing.T, inputAssignments ...[]fr.Element) {
		testMimc(t, numRounds, inputAssignments...)
	}
}

func testMimc(t *testing.T, numRounds int, inputAssignments ...[]fr.Element) {
	//TODO: Implement mimc correctly. Currently, the computation is mimc(a,b) = cipher( cipher( ... cipher(a, b), b) ..., b)
	// @AlexandreBelling: Please explain the extra layers in https://github.com/ConsenSys/gkr-mimc/blob/81eada039ab4ed403b7726b535adb63026e8011f/examples/mimc.go#L10

	c := make(Circuit, numRounds+1)

	c[numRounds] = CircuitLayer{
		{
			Inputs:     []*Wire{},
			NumOutputs: 1,
			Gate:       nil,
		},
		{
			Inputs:     []*Wire{},
			NumOutputs: numRounds,
			Gate:       nil,
		},
	}

	for i := numRounds; i > 0; i-- {
		c[i-1] = CircuitLayer{
			{
				Inputs:     []*Wire{&c[i][0], &c[numRounds][1]},
				NumOutputs: 1,
				Gate:       mimcCipherGate{}, //TODO: Put arks in there
			},
		}
	}

	t.Log("Evaluating all circuit wires")
	assignment := WireAssignment{&c[numRounds][0]: inputAssignments[0], &c[numRounds][1]: inputAssignments[1]}.complete(c)
	t.Log("Circuit evaluation complete")

	proof := Prove(c, assignment, sumcheck.NewMessageCounter(0, 1))

	t.Log("Proof finished")
	if !Verify(c, assignment, proof, sumcheck.NewMessageCounter(0, 1)) {
		t.Error("Proof rejected")
	}

	t.Log("Successful verification finished")
	if Verify(c, assignment, proof, sumcheck.NewMessageCounter(1, 1)) {
		t.Error("Bad proof accepted")
	}
	t.Log("Unsuccessful verification finished")
}

func testATimesBSquared(t *testing.T, numRounds int, inputAssignments ...[]fr.Element) {
	// This imitates the MiMC circuit

	c := make(Circuit, numRounds+1)

	c[numRounds] = CircuitLayer{
		{
			Inputs:     []*Wire{},
			NumOutputs: 1,
			Gate:       nil,
		},
		{
			Inputs:     []*Wire{},
			NumOutputs: numRounds,
			Gate:       nil,
		},
	}

	for i := numRounds; i > 0; i-- {
		c[i-1] = CircuitLayer{
			{
				Inputs:     []*Wire{&c[i][0], &c[numRounds][1]},
				NumOutputs: 1,
				Gate:       mulGate{},
			},
		}
	}

	assignment := WireAssignment{&c[numRounds][0]: inputAssignments[0], &c[numRounds][1]: inputAssignments[1]}.complete(c)

	proof := Prove(c, assignment, sumcheck.NewMessageCounter(0, 1))

	if !Verify(c, assignment, proof, sumcheck.NewMessageCounter(0, 1)) {
		t.Error("Proof rejected")
	}

	if Verify(c, assignment, proof, sumcheck.NewMessageCounter(1, 1)) {
		t.Error("Bad proof accepted")
	}
}

func setRandom(slice []fr.Element) {
	for i := range slice {
		slice[i].SetRandom()
	}
}

type mulGate struct{}

func (m mulGate) Evaluate(element ...fr.Element) (result fr.Element) {
	result.Mul(&element[0], &element[1])
	return
}

func (m mulGate) Degree() int {
	return 2
}

func generateTestProver(path string) func(t *testing.T) {
	return func(t *testing.T) {
		testCase := newTestCase(t, path)
		testCase.Transcript.Update(0)
		proof := Prove(testCase.Circuit, testCase.FullAssignment, testCase.Transcript)
		assertProofEquals(t, testCase.Proof, proof)
	}
}

func generateTestVerifier(path string) func(t *testing.T) {
	return func(t *testing.T) {
		testCase := newTestCase(t, path)
		testCase.Transcript.Update(0)
		success := Verify(testCase.Circuit, testCase.InOutAssignment, testCase.Proof, testCase.Transcript)
		assert.True(t, success)

		testCase = newTestCase(t, path)
		testCase.Transcript.Update(1)
		success = Verify(testCase.Circuit, testCase.InOutAssignment, testCase.Proof, testCase.Transcript)
		assert.False(t, success)
	}
}

func TestGkrVectors(t *testing.T) {

	testDirPath := "../../rational_cases"
	dirEntries, err := os.ReadDir(testDirPath)
	if err != nil {
		t.Error(err)
	}
	for _, dirEntry := range dirEntries {
		if !dirEntry.IsDir() {

			if filepath.Ext(dirEntry.Name()) == ".json" {
				path := filepath.Join(testDirPath, dirEntry.Name())
				noExt := dirEntry.Name()[:len(dirEntry.Name())-len(".json")]

				t.Run(noExt+"_prover", generateTestProver(path))
				t.Run(noExt+"_verifier", generateTestVerifier(path))

			}
		}
	}
}

func TestTestHash(t *testing.T) {
	m := getHash(t, "../../rational_cases/resources/hash.json")
	var one, two, negFour fr.Element
	one.SetOne()
	two.SetInt64(2)
	negFour.SetInt64(-4)

	h := m.hash(&one, &two)
	assert.True(t, h.Equal(&negFour), "expected -4, saw %s", h.Text(10))
}

type TestCase struct {
	Circuit         Circuit
	Transcript      sumcheck.ArithmeticTranscript
	Proof           Proof
	FullAssignment  WireAssignment
	InOutAssignment WireAssignment
}

type TestCaseInfo struct {
	Hash    string          `json:"hash"`
	Circuit string          `json:"circuit"`
	Input   [][]interface{} `json:"input"`
	Output  [][]interface{} `json:"output"`
	Proof   PrintableProof  `json:"proof"`
}

type ParsedTestCase struct {
	FullAssignment  WireAssignment
	InOutAssignment WireAssignment
	Proof           Proof
	Hash            HashMap
	Circuit         Circuit
}

var parsedTestCases = make(map[string]*ParsedTestCase)

func newTestCase(t *testing.T, path string) *TestCase {
	path, err := filepath.Abs(path)
	assert.NoError(t, err)
	dir := filepath.Dir(path)

	parsedCase, ok := parsedTestCases[path]
	if !ok {
		if bytes, err := os.ReadFile(path); err == nil {
			var info TestCaseInfo
			err = json.Unmarshal(bytes, &info)
			if err != nil {
				t.Error(err)
			}

			circuit := getCircuit(t, filepath.Join(dir, info.Circuit))
			hash := getHash(t, filepath.Join(dir, info.Hash))
			proof := unmarshalProof(t, info.Proof)

			fullAssignment := make(WireAssignment)
			inOutAssignment := make(WireAssignment)
			assignmentSize := len(info.Input[0])

			{
				i := len(circuit) - 1

				assert.Equal(t, len(circuit[i]), len(info.Input), "Input layer not the same size as input vector")

				for j := range circuit[i] {
					wire := &circuit[i][j]
					wireAssignment := sliceToElementSlice(t, info.Input[j])
					fullAssignment[wire] = wireAssignment
					inOutAssignment[wire] = wireAssignment
				}
			}

			for i := len(circuit) - 2; i >= 0; i-- {
				for j := range circuit[i] {
					wire := &circuit[i][j]
					assignment := make(polynomial.MultiLin, assignmentSize)
					in := make([]fr.Element, len(wire.Inputs))
					for k := range assignment {
						for l, inputWire := range circuit[i][j].Inputs {
							in[l] = fullAssignment[inputWire][k]
						}
						assignment[k] = wire.Gate.Evaluate(in...)
					}

					fullAssignment[wire] = assignment
				}
			}

			assert.Equal(t, len(circuit[0]), len(info.Output), "Output layer not the same size as output vector")
			for j := range circuit[0] {
				wire := &circuit[0][j]
				inOutAssignment[wire] = sliceToElementSlice(t, info.Output[j])
				assert.NoError(t, sliceEquals(inOutAssignment[wire], fullAssignment[wire]), "circuit output mismatch on wire 0,%d", j)
			}

			parsedCase = &ParsedTestCase{
				FullAssignment:  fullAssignment,
				InOutAssignment: inOutAssignment,
				Proof:           proof,
				Hash:            hash,
				Circuit:         circuit,
			}

			parsedTestCases[path] = parsedCase
		} else {
			t.Error(err)
		}
	}

	return &TestCase{
		Circuit:         parsedCase.Circuit,
		Transcript:      &MapHashTranscript{hashMap: parsedCase.Hash},
		FullAssignment:  parsedCase.FullAssignment,
		InOutAssignment: parsedCase.InOutAssignment,
		Proof:           parsedCase.Proof,
	}
}

type WireInfo struct {
	Gate   string  `json:"gate"`
	Inputs [][]int `json:"inputs"`
}

type CircuitInfo [][]WireInfo

var circuitCache = make(map[string]Circuit)

func getCircuit(t *testing.T, path string) Circuit {
	path, err := filepath.Abs(path)
	if err != nil {
		t.Error(err)
	}
	if circuit, ok := circuitCache[path]; ok {
		return circuit
	}
	if bytes, err := os.ReadFile(path); err == nil {
		var circuitInfo CircuitInfo
		if err := json.Unmarshal(bytes, &circuitInfo); err == nil {
			circuit := circuitInfo.toCircuit()
			circuitCache[path] = circuit
			return circuit
		} else {
			t.Error(err)
		}
	} else {
		t.Error(err)
	}
	return nil //unreachable
}

func (c CircuitInfo) toCircuit() (circuit Circuit) {
	isOutput := make(map[*Wire]interface{})
	circuit = make(Circuit, len(c))
	for i := len(c) - 1; i >= 0; i-- {
		circuit[i] = make(CircuitLayer, len(c[i]))
		for j, wireInfo := range c[i] {
			circuit[i][j].Gate = gates[wireInfo.Gate]
			circuit[i][j].Inputs = make([]*Wire, len(wireInfo.Inputs))
			isOutput[&circuit[i][j]] = nil
			for k, inputCoord := range wireInfo.Inputs {
				if len(inputCoord) != 2 {
					panic("circuit wire has two coordinates")
				}
				input := &circuit[inputCoord[0]][inputCoord[1]]
				input.NumOutputs++
				circuit[i][j].Inputs[k] = input
				delete(isOutput, input)
			}
			if (i == len(c)-1) != (len(circuit[i][j].Inputs) == 0) {
				panic("wire is input if and only if in last layer")
			}
		}
	}

	for k := range isOutput {
		k.NumOutputs = 1
	}

	return
}

var gates map[string]Gate

func init() {
	gates = make(map[string]Gate)
	gates["identity"] = identityGate{}
	gates["mul"] = mulGate{}
	gates["mimc"] = mimcCipherGate{} //TODO: Add ark
}

type mimcCipherGate struct {
	ark fr.Element
}

func (m mimcCipherGate) Evaluate(input ...fr.Element) (res fr.Element) {
	var sum fr.Element

	sum.
		Add(&input[0], &input[1]).
		Add(&sum, &m.ark)

	res.Square(&sum)    // sum^2
	res.Mul(&res, &sum) // sum^3
	res.Square(&res)    //sum^6
	res.Mul(&res, &sum) //sum^7

	return
}

func (m mimcCipherGate) Degree() int {
	return 7
}

type RationalTriplet struct {
	key1        fr.Element
	key2        fr.Element
	key2Present bool
	value       fr.Element
}

func (t *RationalTriplet) CmpKey(o *RationalTriplet) int {
	if cmp1 := t.key1.Cmp(&o.key1); cmp1 != 0 {
		return cmp1
	}

	if t.key2Present {
		if o.key2Present {
			return t.key2.Cmp(&o.key2)
		}
		return 1
	} else {
		if o.key2Present {
			return -1
		}
		return 0
	}
}

var hashCache = make(map[string]HashMap)

func getHash(t *testing.T, path string) HashMap {
	path, err := filepath.Abs(path)
	if err != nil {
		t.Error(err)
	}
	if h, ok := hashCache[path]; ok {
		return h
	}
	if bytes, err := os.ReadFile(path); err == nil {
		var asMap map[string]interface{}
		if err := json.Unmarshal(bytes, &asMap); err != nil {
			t.Error(err)
		}

		res := make(HashMap, 0, len(asMap))

		for k, v := range asMap {
			var entry RationalTriplet
			if _, err := entry.value.Set(v); err != nil {
				t.Error(err)
			}

			key := strings.Split(k, ",")

			switch len(key) {
			case 1:
				entry.key2Present = false
			case 2:
				entry.key2Present = true
				if _, err := entry.key2.Set(key[1]); err != nil {
					t.Error(err)
				}
			default:
				t.Errorf("cannot parse %T as one or two field elements", v)
			}
			if _, err := entry.key1.Set(key[0]); err != nil {
				t.Error(err)
			}

			res = append(res, &entry)
		}

		sort.Slice(res, func(i, j int) bool {
			return res[i].CmpKey(res[j]) <= 0
		})

		hashCache[path] = res

		return res

	} else {
		t.Error(err)
	}
	return nil //Unreachable
}

type HashMap []*RationalTriplet

type MapHashTranscript struct {
	hashMap         HashMap
	stateValid      bool
	resultAvailable bool
	state           fr.Element
}

func (m HashMap) hash(x *fr.Element, y *fr.Element) fr.Element {

	toFind := RationalTriplet{
		key1:        *x,
		key2Present: y != nil,
	}

	if y != nil {
		toFind.key2 = *y
	}

	i := sort.Search(len(m), func(i int) bool { return m[i].CmpKey(&toFind) >= 0 })

	if i < len(m) && m[i].CmpKey(&toFind) == 0 {
		return m[i].value
	}

	if y == nil {
		panic("No hash available for input " + x.Text(10))
	} else {
		panic("No hash available for input " + x.Text(10) + "," + y.Text(10))
	}
}

func (m *MapHashTranscript) Update(i ...interface{}) {
	if len(i) > 0 {
		for _, x := range i {

			var xElement fr.Element
			if _, err := xElement.Set(x); err != nil {
				panic(err.Error())
			}
			if m.stateValid {
				m.state = m.hashMap.hash(&xElement, &m.state)
			} else {
				m.state = m.hashMap.hash(&xElement, nil)
			}

			m.stateValid = true
		}
	} else { //just hash the state itself
		if !m.stateValid {
			panic("nothing to hash")
		}
		m.state = m.hashMap.hash(&m.state, nil)
	}
	m.resultAvailable = true
}

func (m *MapHashTranscript) Next(i ...interface{}) fr.Element {

	if len(i) > 0 || !m.resultAvailable {
		m.Update(i...)
	}
	m.resultAvailable = false
	return m.state
}

func (m *MapHashTranscript) NextN(N int, i ...interface{}) []fr.Element {

	if len(i) > 0 {
		m.Update(i...)
	}

	res := make([]fr.Element, N)

	for n := range res {
		res[n] = m.Next()
	}

	return res
}

func sliceToElementSlice(t *testing.T, slice []interface{}) (elementSlice []fr.Element) {
	elementSlice = make([]fr.Element, len(slice))
	for i, v := range slice {
		if _, err := elementSlice[i].Set(v); err != nil {
			t.Error(err)
		}
	}
	return
}

func sliceEquals(a []fr.Element, b []fr.Element) error {
	if len(a) != len(b) {
		return fmt.Errorf("length mismatch %d≠%d", len(a), len(b))
	}
	for i := range a {
		if !a[i].Equal(&b[i]) {
			return fmt.Errorf("at index %d: %s ≠ %s", i, a[i].String(), b[i].String())
		}
	}
	return nil
}

func assertProofEquals(t *testing.T, expected Proof, seen Proof) {
	assert.Equal(t, len(expected), len(seen))
	for i, x := range expected {
		xSeen := seen[i]
		assert.Equal(t, len(x), len(xSeen))
		for j, y := range x {
			ySeen := xSeen[j]

			if ySeen.FinalEvalProof == nil {
				assert.Equal(t, 0, len(y.FinalEvalProof.([]fr.Element)))
			} else {
				assert.Equal(t, y.FinalEvalProof, ySeen.FinalEvalProof)
			}
			assert.Equal(t, len(y.PartialSumPolys), len(ySeen.PartialSumPolys))
			for k, z := range y.PartialSumPolys {
				zSeen := ySeen.PartialSumPolys[k]
				assert.NoError(t, sliceEquals(z, zSeen))
			}
		}
	}
}

type PrintableProof [][]PrintableSumcheckProof

type PrintableSumcheckProof struct {
	FinalEvalProof  interface{}     `json:"finalEvalProof"`
	PartialSumPolys [][]interface{} `json:"partialSumPolys"`
}

func unmarshalProof(t *testing.T, printable PrintableProof) (proof Proof) {
	proof = make(Proof, len(printable))
	for i := range printable {
		proof[i] = make([]sumcheck.Proof, len(printable[i]))
		for j, printableSumcheck := range printable[i] {
			finalEvalProof := []fr.Element(nil)

			if printableSumcheck.FinalEvalProof != nil {
				finalEvalSlice := reflect.ValueOf(printableSumcheck.FinalEvalProof)
				finalEvalProof = make([]fr.Element, finalEvalSlice.Len())
				for k := range finalEvalProof {
					_, err := finalEvalProof[k].Set(finalEvalSlice.Index(k).Interface())
					assert.NoError(t, err)
				}
			}

			proof[i][j] = sumcheck.Proof{
				PartialSumPolys: make([]polynomial.Polynomial, len(printableSumcheck.PartialSumPolys)),
				FinalEvalProof:  finalEvalProof,
			}
			for k := range printableSumcheck.PartialSumPolys {
				proof[i][j].PartialSumPolys[k] = sliceToElementSlice(t, printableSumcheck.PartialSumPolys[k])
			}
		}
	}
	return
}
