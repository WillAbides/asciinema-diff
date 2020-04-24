package diff

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_compareHeaders(t *testing.T) {
	for _, td := range []struct {
		name   string
		a      map[string]interface{}
		b      map[string]interface{}
		fields []string
		want   bool
	}{
		{
			name: "both nil",
			want: true,
		},
		{
			name: "equality where it matters",
			a: map[string]interface{}{
				"foo": "bar",
				"bar": "asdf",
				"env": map[string]interface{}{
					"a": "b",
					"c": "d",
				},
			},
			b: map[string]interface{}{
				"foo": "bar",
				"bar": "jkl;",
				"env": map[string]interface{}{
					"a": "b",
					"c": "d",
				},
			},
			fields: []string{"foo", "notpresent", "env"},
			want:   true,
		},
	} {
		t.Run(td.name, func(t *testing.T) {
			got := compareHeaders(td.a, td.b, td.fields...)
			require.Equal(t, td.want, got)
		})
	}
}

func TestEqual(t *testing.T) {
	for _, td := range []struct {
		name    string
		aFile   string
		bFile   string
		options []EqualOption
		want    bool
		wantErr bool
	}{
		{
			name:    "all good",
			aFile:   "testdata/foo.cast",
			bFile:   "testdata/bar.cast",
			options: []EqualOption{TimeTolerance(50 * time.Millisecond)},
			want:    true,
		},
		{
			name:  "no tolerance",
			aFile: "testdata/foo.cast",
			bFile: "testdata/bar.cast",
			want:  false,
		},
		{
			name:  "a is longer",
			aFile: "testdata/foo.cast",
			bFile: "testdata/foo-short.cast",
			want:  false,
		},
		{
			name:  "b is longer",
			aFile: "testdata/foo-short.cast",
			bFile: "testdata/foo.cast",
			want:  false,
		},
		{
			name:  "different headers",
			aFile: "testdata/foo.cast",
			bFile: "testdata/bar.cast",
			options: []EqualOption{
				TimeTolerance(50 * time.Millisecond),
				CompareHeaderFields("timestamp"),
			},
			want: false,
		},
	} {
		t.Run(td.name, func(t *testing.T) {
			aRdr, err := os.Open(filepath.FromSlash(td.aFile))
			require.NoError(t, err)
			t.Cleanup(func() {
				require.NoError(t, aRdr.Close())
			})
			bRdr, err := os.Open(filepath.FromSlash(td.bFile))
			require.NoError(t, err)
			t.Cleanup(func() {
				require.NoError(t, bRdr.Close())
			})
			got, err := Equal(aRdr, bRdr, td.options...)
			if td.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, td.want, got)
		})
	}
}

func Test_event_equal(t *testing.T) {
	for _, td := range []struct {
		name      string
		a         *event
		b         *event
		tolerance time.Duration
		want      bool
	}{
		{
			name: "exact match",
			a: &event{
				Time: 1000,
				Type: "o",
				Data: "foo",
			},
			b: &event{
				Time: 1000,
				Type: "o",
				Data: "foo",
			},
			tolerance: 0,
			want:      true,
		},
		{
			name: "data mismatch",
			a: &event{
				Time: 1000,
				Type: "o",
				Data: "foo",
			},
			b: &event{
				Time: 1000,
				Type: "o",
				Data: "bar",
			},
			tolerance: 0,
			want:      false,
		},
		{
			name: "type mismatch",
			a: &event{
				Time: 1000,
				Type: "o",
				Data: "foo",
			},
			b: &event{
				Time: 1000,
				Type: "i",
				Data: "foo",
			},
			tolerance: 0,
			want:      false,
		},
		{
			name: "nil receiver",
			a:    nil,
			b: &event{
				Time: 1000,
				Type: "i",
				Data: "foo",
			},
			tolerance: 0,
			want:      false,
		},
		{
			name: "nil arg",
			b:    nil,
			a: &event{
				Time: 1000,
				Type: "i",
				Data: "foo",
			},
			tolerance: 0,
			want:      false,
		},
		{
			name:      "nil receiver and arg",
			b:         nil,
			a:         nil,
			tolerance: 0,
			want:      true,
		},
		{
			name: "within tolerance",
			a: &event{
				Time: 1000,
				Type: "o",
				Data: "foo",
			},
			b: &event{
				Time: 1010,
				Type: "o",
				Data: "foo",
			},
			tolerance: 100,
			want:      true,
		},
		{
			name: "above tolerance",
			a: &event{
				Time: 1000,
				Type: "o",
				Data: "foo",
			},
			b: &event{
				Time: 1110,
				Type: "o",
				Data: "foo",
			},
			tolerance: 100,
			want:      false,
		},
		{
			name: "below tolerance",
			a: &event{
				Time: 1000,
				Type: "o",
				Data: "foo",
			},
			b: &event{
				Time: 880,
				Type: "o",
				Data: "foo",
			},
			tolerance: 100,
			want:      false,
		},
	} {
		t.Run(td.name, func(t *testing.T) {
			got := td.a.equal(td.b, td.tolerance)
			require.Equal(t, td.want, got)
		})
	}
}

func Test_event_UnmarshalJSON(t *testing.T) {
	data := []byte(`[1.184334, "o", "\u001b[?1034h$ "]`)
	want := &event{
		Time: time.Duration(1.184334 * float64(time.Second)),
		Type: "o",
		Data: "\u001b[?1034h$ ",
	}
	ev := &event{}
	err := ev.UnmarshalJSON(data)
	require.NoError(t, err)
	require.Equal(t, 1.184334, ev.Time.Seconds())
	require.Equal(t, want, ev)
}
