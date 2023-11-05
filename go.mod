module haxsh

go 1.21.3

// there's a non-`require` main dependency: the seperate repo `yo`, which becomes available by
// having it as a sibling-to-this-repo dir; plus, one level up, the following `go.work` file:

// go 1.21.3
// use ./yo
// use ./haxsh
