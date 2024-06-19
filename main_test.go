package main

import "testing"

var cases = []struct {
	filter string
	name   string
	want   bool
}{
	{
		filter: "foo",
		name:   "foo",
		want:   true,
	},
	{
		filter: "foo",
		name:   "bar",
		want:   false,
	},
	{
		filter: "foo/bar",
		name:   "foo/bar",
		want:   true,
	},
	{
		filter: "foo/bar",
		name:   "foo/baz",
		want:   false,
	},
	{
		filter: "foo/+",
		name:   "foo/bar",
		want:   true,
	},
	{
		filter: "foo/+",
		name:   "foo",
		want:   false,
	},
	{
		filter: "foo/+",
		name:   "foo/",
		want:   true,
	},
	{
		filter: "foo/+/bar",
		name:   "foo/baz/bar",
		want:   true,
	},
	{
		filter: "foo/+/bar",
		name:   "foo//bar",
		want:   true,
	},
	{
		filter: "foo/#",
		name:   "foo/bar/baz",
		want:   true,
	},
	{
		filter: "foo/#",
		name:   "foo/",
		want:   true,
	},
	{
		filter: "foo/#",
		name:   "foo",
		want:   false,
	},
	{
		filter: "foo/+/bar/#",
		name:   "foo/baz/bar/qux",
		want:   true,
	},
}

func TestTopicMatching(t *testing.T) {
	for _, c := range cases {
		got := topicMatches(c.filter, c.name)
		if got != c.want {
			t.Fatalf(
				`wanted %t but got %t, topicMatches("%s", "%s")`,
				c.want,
				got,
				c.filter,
				c.name,
			)
		}
	}
}

func BenchmarkTopicMatching(b *testing.B) {
	for i := range b.N {
		i = i % len(cases)

		got := topicMatches(cases[i].filter, cases[i].name)

		if got != cases[i].want {
			b.Fatalf(
				`wanted %t but got %t, topicMatches("%s", "%s")`,
				cases[i].want,
				got,
				cases[i].filter,
				cases[i].name,
			)
		}
	}
}
