package main

import (
	internalApp "github.com/billgoswell/commandlinetodo/internal/app"
)

// Re-export constants and types from internal/app for backward compatibility
type AppState = internalApp.AppState
type SubState = internalApp.SubState

const (
	InputModeTask     = internalApp.InputModeTask
	InputModePriority = internalApp.InputModePriority
	InputModeDueDate  = internalApp.InputModeDueDate
	InputModeListName = internalApp.InputModeListName

	StateMainBrowse        = internalApp.StateMainBrowse
	StateTaskInput         = internalApp.StateTaskInput
	StatePrioritySelection = internalApp.StatePrioritySelection
	StateDueDateInput      = internalApp.StateDueDateInput
	StateEditTask          = internalApp.StateEditTask
	StateDeleteConfirm     = internalApp.StateDeleteConfirm
	StateListSelector      = internalApp.StateListSelector
	StateListNameInput     = internalApp.StateListNameInput

	SubStateNone              = internalApp.SubStateNone
	SubStateListManage        = internalApp.SubStateListManage
	SubStateListDeleteConfirm = internalApp.SubStateListDeleteConfirm
	SubStateEditDueDate       = internalApp.SubStateEditDueDate
	SubStateListRename        = internalApp.SubStateListRename
	SubStateListCreate        = internalApp.SubStateListCreate

	PriorityHigh    = internalApp.PriorityHigh
	PriorityMedHigh = internalApp.PriorityMedHigh
	PriorityMed     = internalApp.PriorityMed
	PriorityLow     = internalApp.PriorityLow

	KeyUp    = internalApp.KeyUp
	KeyDown  = internalApp.KeyDown
	KeyEnter = internalApp.KeyEnter
	KeyEsc   = internalApp.KeyEsc
	KeyQ     = internalApp.KeyQ
	KeyCtrlC = internalApp.KeyCtrlC
	KeySpace = internalApp.KeySpace
	KeyK     = internalApp.KeyK
	KeyJ     = internalApp.KeyJ
	KeyA     = internalApp.KeyA
	KeyE     = internalApp.KeyE
	KeyD     = internalApp.KeyD
	KeyT     = internalApp.KeyT
	KeyY     = internalApp.KeyY
	KeyN     = internalApp.KeyN
	KeyW     = internalApp.KeyW
	KeyR     = internalApp.KeyR
	KeyL     = internalApp.KeyL
	KeyM     = internalApp.KeyM
	KeyS     = internalApp.KeyS

	KeyPriority1 = internalApp.KeyPriority1
	KeyPriority2 = internalApp.KeyPriority2
	KeyPriority3 = internalApp.KeyPriority3
	KeyPriority4 = internalApp.KeyPriority4

	ViewportWidth     = internalApp.ViewportWidth
	ViewportHeight    = internalApp.ViewportHeight
	ViewportBaseLines = internalApp.ViewportBaseLines
	LinesPerTask      = internalApp.LinesPerTask

	TextInputCharLimit = internalApp.TextInputCharLimit
	TextInputWidth     = internalApp.TextInputWidth

	DefaultPriority = internalApp.DefaultPriority
	DefaultListName = internalApp.DefaultListName

	TextInputPlaceholder = internalApp.TextInputPlaceholder

	SecondsPerDay = internalApp.SecondsPerDay
	MaxDaysOffset = internalApp.MaxDaysOffset
)

var (
	TitleStyle         = internalApp.TitleStyle
	SelectedStyle      = internalApp.SelectedStyle
	DoneStyle          = internalApp.DoneStyle
	SelectedDoneStyle  = internalApp.SelectedDoneStyle
	ItemStyle          = internalApp.ItemStyle
	DocStyle           = internalApp.DocStyle
	OneStyle           = internalApp.OneStyle
	TwoStyle           = internalApp.TwoStyle
	ThreeStyle         = internalApp.ThreeStyle
	FourStyle          = internalApp.FourStyle
	SelectedOneStyle   = internalApp.SelectedOneStyle
	SelectedTwoStyle   = internalApp.SelectedTwoStyle
	SelectedThreeStyle = internalApp.SelectedThreeStyle
	SelectedFourStyle  = internalApp.SelectedFourStyle
	OverdueStyle       = internalApp.OverdueStyle
	ErrorStyle         = internalApp.ErrorStyle
	SyncingStyle       = internalApp.SyncingStyle
	SuccessStyle       = internalApp.SuccessStyle
	OfflineStyle       = internalApp.OfflineStyle

	SpinnerFrames              = internalApp.SpinnerFrames
	PriorityStyles             = internalApp.PriorityStyles
	PriorityLabels             = internalApp.PriorityLabels
	SelectedPriorityStyles     = internalApp.SelectedPriorityStyles
	StateHeightAdjustments     = internalApp.StateHeightAdjustments
)
