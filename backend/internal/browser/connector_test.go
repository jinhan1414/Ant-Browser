package browser

import (
	"reflect"
	"testing"
)

func TestBuildLaunchArgsKeepsOriginalArgs(t *testing.T) {
	t.Parallel()

	baseArgs := []string{"--disable-sync"}
	got := BuildLaunchArgs(append([]string{}, baseArgs...), &Profile{})
	want := []string{"--disable-sync"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("BuildLaunchArgs 结果错误:\n got=%v\nwant=%v", got, want)
	}
}
