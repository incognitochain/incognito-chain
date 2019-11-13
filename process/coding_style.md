# Incognito Coding Style

Incognito, at over 1 million lines of code and 25+ fulltime developers, is one of the largest blockchain codebase. It's crucial to have a guideline for existing developers to work together. It's also helpful for new developers wishing to participate in its development.

## Effective Go

Read [Effective Go](https://golang.org/doc/effective_go.html). The majority of Incognito codebase is written in Go. Effective Go is a great starting point to learn how to write idiomatic Go code.

## Go Code Review Comments

Read [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments). It is a good supplement to Effective Go.

## Logging

Only use log.Logger to write logs.  Do not use fmt.

## Constructors

Use **New** prefix for all constructor functions.

```go
func NewShardBlock() *ShardBlock {
        return &ShardBlock{
                ...
        }
}
```

## Function params

Always name function params.

<img src="https://i.postimg.cc/dQf6ZL2L/768px-No-icon-red-svg.png" width=50>

```go
AddShardRewardRequest(uint64, byte, uint64, common.Hash) error
```

<img src="https://i.postimg.cc/JhJVvfRQ/check-checkbox-checkmark-confirm-success-yes-icon-13201967112260.png" width=50>

```go
AddShardRewardRequest(epoch uint64, shardID byte, amount uint64, tokenID common.Hash) error
```

