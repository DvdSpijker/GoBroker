package main

import (
  "time"
)

type (
  Color struct {
    Red byte
    Green byte
    Blue byte
  }

  Gradient struct {
    Red int8
    Green int8
    Blue int8
  }

  Leaf struct {
    Index int

    Color Color
    Gradient Gradient

    Brightness int
    FadeCoefficient float32
  }

  LeafRow []Leaf

  Season struct {
    End time.Time

    Colors []Color
    MaxGradient Gradient
    MinGradient Gradient

    BeginBrightess int
    FinalBrigthness int
    MaxFade float32
    MinFade float32
  }
)

var (
  Trunk []int = ConsecutiveNumbers(0, 35)
  MainBranch []int = ConsecutiveNumbers(36, 46)
  Branches []int = Merge(
    ConsecutiveNumbers(96,99),
    ConsecutiveNumbers(113,115),
    ConsecutiveNumbers(128,130),
    ConsecutiveNumbers(144, 148))
  Outline []int = ConsecutiveNumbers(47, 89)
  LeftLeafRow0 []int = ConsecutiveNumbers(90, 94)
  LeftLeafRow1 []int = []int{100}
  LeftLeafRow2 []int = ConsecutiveNumbers(101, 106)
  LeftLeafRow3 []int = ConsecutiveNumbers(107, 112)
  LeftLeafRow4 []int = []int{116}
  LeftLeafRow5 []int = ConsecutiveNumbers(117,119)
  LeftLeafRow6 []int = []int{120,121}
  Peak []int = []int{122}
  RightLeafRow6 []int = []int{123,124}
  RightLeafRow5 []int = ConsecutiveNumbers(125,126)
  RightLeafRow4 []int = []int{127}
  RightLeafRow3 []int = ConsecutiveNumbers(131,136)
  RightLeafRow2 []int = ConsecutiveNumbers(137,142)
  RightLeafRow1 []int = []int{143}
  RightLeafRow0 []int = ConsecutiveNumbers(149,153)

  Spring Season = Season{}
  Summer Season = Season{}
  Fall Season = Season{}
  Winter Season = Season{}

  Year []Season =[]Season{
    Spring,
    Summer,
    Fall,
    Winter,
  }
)

func Leafs() []int {
  return Merge(
    LeftLeafRow0,
    LeftLeafRow1,
    LeftLeafRow2,
    LeftLeafRow3,
    LeftLeafRow4,
    LeftLeafRow5,
    LeftLeafRow6,
    Peak,
    RightLeafRow0,
    RightLeafRow1,
    RightLeafRow2,
    RightLeafRow3,
    RightLeafRow4,
    RightLeafRow5,
    RightLeafRow6,
    )
}

func Tree() []int {
  return ConsecutiveNumbers(0, 153)
}

func ConsecutiveNumbers(start, end int) []int {
  numbers := make([]int, end-start+1)
  for i := 0; i < len(numbers); i++ {
    numbers[i] = start+i
  }
  return numbers
} 

func Merge(ranges... []int) []int {
  total := 0
  for _, r := range ranges {
    total += len(r)
  }
  result := make([]int, 0, total)
  for _, r := range ranges {
    result = append(result, r...)
  }

  return result
}
