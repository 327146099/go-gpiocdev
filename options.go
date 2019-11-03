// SPDX-License-Identifier: MIT
//
// Copyright © 2019 Kent Gibson <warthog618@gmail.com>.

// +build linux

package gpiod

import "github.com/warthog618/gpiod/uapi"

// ChipOption defines the interface required to provide a Chip option.
type ChipOption interface {
	applyChipOption(*ChipOptions)
}

// ChipOptions contains the options for a Chip.
type ChipOptions struct {
	consumer string
}

// ConsumerOption defines the consumer label for a line.
type ConsumerOption string

// WithConsumer provides the consumer label for the line.
//
// When applied to a chip it provides the default consumer label for all lines
// requested by the chip.
func WithConsumer(consumer string) ConsumerOption {
	return ConsumerOption(consumer)

}

func (o ConsumerOption) applyChipOption(c *ChipOptions) {
	c.consumer = string(o)
}

func (o ConsumerOption) applyLineOption(l *LineOptions) {
	l.consumer = string(o)
}

// LineOption defines the interface required to provide an option for Line and
// Lines.
type LineOption interface {
	applyLineOption(*LineOptions)
}

// LineConfig defines the interface required to update an option for Line and
// Lines.
type LineConfig interface {
	applyLineConfig(*LineOptions)
}

// LineOptions contains the options for a Line or Lines.
type LineOptions struct {
	consumer      string
	InitialValues []int
	EventFlags    uapi.EventFlag
	HandleFlags   uapi.HandleFlag
	eh            EventHandler
}

// EventHandler is a receiver for line events.
type EventHandler func(LineEvent)

// AsIsOption indicates the line direction should be left as is.
type AsIsOption struct{}

// AsIs indicates that a line be requested as neither an input or output.
//
// That is its direction is left as is. This option overrides and clears any
// previous Input or Output options.
var AsIs = AsIsOption{}

func (o AsIsOption) applyLineOption(l *LineOptions) {
	l.HandleFlags &= ^(uapi.HandleRequestOutput | uapi.HandleRequestInput)
}

// InputOption indicates the line direction should be set to an input.
type InputOption struct{}

// AsInput indicates that a line be requested as an input.
//
// This option overrides and clears any previous Output, OpenDrain, or
// OpenSource options.
var AsInput = InputOption{}

func (o InputOption) applyLineOption(l *LineOptions) {
	l.HandleFlags &= ^(uapi.HandleRequestOutput |
		uapi.HandleRequestOpenDrain |
		uapi.HandleRequestOpenSource)
	l.HandleFlags |= uapi.HandleRequestInput
}

func (o InputOption) applyLineConfig(l *LineOptions) {
	o.applyLineOption(l)
}

// OutputOption indicates the line direction should be set to an output.
type OutputOption struct {
	initialValues []int
}

// AsOutput indicates that a line or lines be requested as an output.
//
// The initial active state for the line(s) can optionally be provided.
// If fewer values are provided than lines then the remaining lines default to
// inactive.
//
// This option overrides and clears any previous Input, RisingEdge, FallingEdge,
// or BothEdges options.
func AsOutput(values ...int) OutputOption {
	vv := append([]int(nil), values...)
	return OutputOption{vv}
}

func (o OutputOption) applyLineOption(l *LineOptions) {
	l.HandleFlags &= ^uapi.HandleRequestInput
	l.HandleFlags |= uapi.HandleRequestOutput
	l.EventFlags = 0
	l.InitialValues = o.initialValues
}

func (o OutputOption) applyLineConfig(l *LineOptions) {
	o.applyLineOption(l)
}

// FlagOption applies particular line handle flags.
type FlagOption struct {
	flag uapi.HandleFlag
}

func (o FlagOption) applyLineOption(l *LineOptions) {
	l.HandleFlags |= o.flag
}

// AsActiveLow indicates that a line be considered active when the line level
// is low.
var AsActiveLow = FlagOption{uapi.HandleRequestActiveLow}

// DriveModeOption indicates that a line is open drain or open source.
type DriveModeOption struct {
	flag uapi.HandleFlag
}

func (o DriveModeOption) applyLineOption(l *LineOptions) {
	l.HandleFlags &= ^(uapi.HandleRequestInput |
		uapi.HandleRequestOpenDrain |
		uapi.HandleRequestOpenSource)
	l.HandleFlags |= (o.flag | uapi.HandleRequestOutput)
	l.EventFlags = 0
}

func (o DriveModeOption) applyLineConfig(l *LineOptions) {
	o.applyLineOption(l)
}

// AsOpenDrain indicates that a line be driven low but left floating for high.
//
// This option sets the Output option and overrides and clears any previous
// Input, RisingEdge, FallingEdge, BothEdges, or OpenSource options.
var AsOpenDrain = DriveModeOption{uapi.HandleRequestOpenDrain}

// AsOpenSource indicates that a line be driven low but left floating for hign.
//
// This option sets the Output option and overrides and clears any previous
// Input, RisingEdge, FallingEdge, BothEdges, or OpenDrain options.
var AsOpenSource = DriveModeOption{uapi.HandleRequestOpenSource}

// BiasModeOption indicates how a line is to be biased.
type BiasModeOption struct {
	flag uapi.HandleFlag
}

func (o BiasModeOption) applyLineOption(l *LineOptions) {
	l.HandleFlags &= ^(uapi.HandleRequestBiasDisable |
		uapi.HandleRequestPullDown |
		uapi.HandleRequestPullUp)
	l.HandleFlags |= o.flag
}

func (o BiasModeOption) applyLineConfig(l *LineOptions) {
	o.applyLineOption(l)
}

// WithBiasDisable indicates that a line have its internal bias disabled.
//
// This option overrides and clears any previous bias options.
var WithBiasDisable = BiasModeOption{uapi.HandleRequestBiasDisable}

// WithPullDown indicates that a line have its internal pull-down enabled.
//
// This option overrides and clears any previous bias options.
var WithPullDown = BiasModeOption{uapi.HandleRequestPullDown}

// WithPullUp indicates that a line have its internal pull-up enabled.
//
// This option overrides and clears any previous bias options.
var WithPullUp = BiasModeOption{uapi.HandleRequestPullUp}

// EdgeOption indicates that a line will generate events when edges are detected.
type EdgeOption struct {
	e    EventHandler
	edge uapi.EventFlag
}

func (o EdgeOption) applyLineOption(l *LineOptions) {
	l.HandleFlags &= ^(uapi.HandleRequestOutput |
		uapi.HandleRequestOpenDrain |
		uapi.HandleRequestOpenSource)
	l.HandleFlags |= uapi.HandleRequestInput
	l.EventFlags = o.edge
	l.eh = o.e
}

// WithFallingEdge indicates that a line will generate events when its active
// state transitions from high to low.
//
// Events are forwarded to the provided handler function.
// This option sets the Input option and overrides and clears any previous
// Output, OpenDrain, or OpenSource options.
func WithFallingEdge(e func(LineEvent)) EdgeOption {
	return EdgeOption{EventHandler(e), uapi.EventRequestFallingEdge}
}

// WithRisingEdge indicates that a line will generate events when its active
// state transitions from low to high.
//
// Events are forwarded to the provided handler function.
// This option sets the Input option and overrides and clears any previous
// Output, OpenDrain, or OpenSource options.
func WithRisingEdge(e func(LineEvent)) EdgeOption {
	return EdgeOption{EventHandler(e), uapi.EventRequestRisingEdge}
}

// WithBothEdges indicates that a line will generate events when its active
// state transitions from low to high and from high to low.
//
// Events are forwarded to the provided handler function.
// This option sets the Input option and overrides and clears any previous
// Output, OpenDrain, or OpenSource options.
func WithBothEdges(e func(LineEvent)) EdgeOption {
	return EdgeOption{EventHandler(e), uapi.EventRequestBothEdges}
}
