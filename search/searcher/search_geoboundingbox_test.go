//  Copyright (c) 2017 Couchbase, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package searcher

import (
	"reflect"
	"testing"

	"github.com/blevesearch/bleve/document"
	"github.com/blevesearch/bleve/index"
	"github.com/blevesearch/bleve/index/store/gtreap"
	"github.com/blevesearch/bleve/index/upsidedown"
	"github.com/blevesearch/bleve/search"
)

func TestGeoBoundingBox(t *testing.T) {

	tests := []struct {
		minLon float64
		minLat float64
		maxLon float64
		maxLat float64
		field  string
		want   []string
	}{
		{10.001, 10.001, 20.002, 20.002, "loc", nil},
		{0.001, 0.001, 0.002, 0.002, "loc", []string{"a"}},
		{0.001, 0.001, 1.002, 1.002, "loc", []string{"a", "b"}},
		{0.001, 0.001, 9.002, 9.002, "loc", []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}},
		// same upper-left, bottom-right point
		{25, 25, 25, 25, "loc", nil},
		// box that would return points, but points reversed
		{0.002, 0.002, 0.001, 0.001, "loc", nil},
	}

	i := setupGeo(t)
	indexReader, err := i.Reader()
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = indexReader.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	for _, test := range tests {
		got, err := testGeoBoundingBoxSearch(indexReader, test.minLon, test.minLat, test.maxLon, test.maxLat, test.field)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("expected %v, got %v for %f %f %f %f %s", test.want, got, test.minLon, test.minLat, test.maxLon, test.maxLat, test.field)
		}

	}
}

func testGeoBoundingBoxSearch(i index.IndexReader, minLon, minLat, maxLon, maxLat float64, field string) ([]string, error) {
	var rv []string
	gbs, err := NewGeoBoundingBoxSearcher(i, minLon, minLat, maxLon, maxLat, field, 1.0, search.SearcherOptions{}, true)
	if err != nil {
		return nil, err
	}
	ctx := &search.SearchContext{
		DocumentMatchPool: search.NewDocumentMatchPool(gbs.DocumentMatchPoolSize(), 0),
	}
	docMatch, err := gbs.Next(ctx)
	for docMatch != nil && err == nil {
		rv = append(rv, string(docMatch.IndexInternalID))
		docMatch, err = gbs.Next(ctx)
	}
	if err != nil {
		return nil, err
	}
	return rv, nil
}

func setupGeo(t *testing.T) index.Index {

	analysisQueue := index.NewAnalysisQueue(1)
	i, err := upsidedown.NewUpsideDownCouch(
		gtreap.Name,
		map[string]interface{}{
			"path": "",
		},
		analysisQueue)
	if err != nil {
		t.Fatal(err)
	}
	err = i.Open()
	if err != nil {
		t.Fatal(err)
	}
	err = i.Update(&document.Document{
		ID: "a",
		Fields: []document.Field{
			document.NewGeoPointField("loc", []uint64{}, 0.0015, 0.0015),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = i.Update(&document.Document{
		ID: "b",
		Fields: []document.Field{
			document.NewGeoPointField("loc", []uint64{}, 1.0015, 1.0015),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = i.Update(&document.Document{
		ID: "c",
		Fields: []document.Field{
			document.NewGeoPointField("loc", []uint64{}, 2.0015, 2.0015),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = i.Update(&document.Document{
		ID: "d",
		Fields: []document.Field{
			document.NewGeoPointField("loc", []uint64{}, 3.0015, 3.0015),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = i.Update(&document.Document{
		ID: "e",
		Fields: []document.Field{
			document.NewGeoPointField("loc", []uint64{}, 4.0015, 4.0015),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = i.Update(&document.Document{
		ID: "f",
		Fields: []document.Field{
			document.NewGeoPointField("loc", []uint64{}, 5.0015, 5.0015),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = i.Update(&document.Document{
		ID: "g",
		Fields: []document.Field{
			document.NewGeoPointField("loc", []uint64{}, 6.0015, 6.0015),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = i.Update(&document.Document{
		ID: "h",
		Fields: []document.Field{
			document.NewGeoPointField("loc", []uint64{}, 7.0015, 7.0015),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = i.Update(&document.Document{
		ID: "i",
		Fields: []document.Field{
			document.NewGeoPointField("loc", []uint64{}, 8.0015, 8.0015),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = i.Update(&document.Document{
		ID: "j",
		Fields: []document.Field{
			document.NewGeoPointField("loc", []uint64{}, 9.0015, 9.0015),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	return i
}

func TestComputeGeoRange(t *testing.T) {
	tests := []struct {
		degs        float64
		maxTerms    int
		onBoundary  int
		offBoundary int
		err         string
	}{
		{0.01, 0, 4, 0, ""},
		{0.1, 0, 56, 144, ""},
		{100.0, 0, 32768, 258560, ""},
		{100.0, 1024, 0, 0, "geo range produces too many terms, so should have error"},
	}

	for _, test := range tests {
		onBoundaryRes, offBoundaryRes, err := ComputeGeoRange(0, GeoBitsShift1Minus1,
			-1.0*test.degs, -1.0*test.degs, test.degs, test.degs, true, test.maxTerms)
		if (err != nil) != (test.err != "") {
			t.Errorf("test: %+v, err: %v", test, err)
		}
		if len(onBoundaryRes) != test.onBoundary {
			t.Errorf("test: %+v, onBoundaryRes: %v", test, len(onBoundaryRes))
		}
		if len(offBoundaryRes) != test.offBoundary {
			t.Errorf("test: %+v, offBoundaryRes: %v", test, len(offBoundaryRes))
		}
	}
}

// --------------------------------------------------------------------

func BenchmarkComputeGeoRangePt01(b *testing.B) {
	onBoundary := 4
	offBoundary := 0
	benchmarkComputeGeoRange(b, -0.01, -0.01, 0.01, 0.01, onBoundary, offBoundary)
}

func BenchmarkComputeGeoRangePt1(b *testing.B) {
	onBoundary := 56
	offBoundary := 144
	benchmarkComputeGeoRange(b, -0.1, -0.1, 0.1, 0.1, onBoundary, offBoundary)
}

func BenchmarkComputeGeoRange10(b *testing.B) {
	onBoundary := 5464
	offBoundary := 53704
	benchmarkComputeGeoRange(b, -10.0, -10.0, 10.0, 10.0, onBoundary, offBoundary)
}

func BenchmarkComputeGeoRange100(b *testing.B) {
	onBoundary := 32768
	offBoundary := 258560
	benchmarkComputeGeoRange(b, -100.0, -100.0, 100.0, 100.0, onBoundary, offBoundary)
}

// --------------------------------------------------------------------

func benchmarkComputeGeoRange(b *testing.B,
	minLon, minLat, maxLon, maxLat float64, onBoundary, offBoundary int) {
	checkBoundaries := true

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		onBoundaryRes, offBoundaryRes, err :=
			ComputeGeoRange(0, GeoBitsShift1Minus1, minLon, minLat, maxLon, maxLat, checkBoundaries, 0)
		if err != nil {
			b.Fatalf("expected no err")
		}
		if len(onBoundaryRes) != onBoundary || len(offBoundaryRes) != offBoundary {
			b.Fatalf("boundaries not matching")
		}
	}
}
