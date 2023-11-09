module kaffe

go 1.21.3

// there's a non-`require` main dependency: the seperate repo `yo`, which becomes available to `gopls` and
// `go <command>` by having it as a sibling-to-this-repo dir; plus, one level up, the following `go.work` file:

// go 1.21.3
// use ./yo
// use ./kaffe

require github.com/NortySpock/eliza-go v0.0.0-20210602021720-7607f5bc3af5
