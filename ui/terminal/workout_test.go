package terminal

import (
	"testing"
	"time"

	"seneschal/config"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestWC() *config.WorkoutConfig {
	return &config.WorkoutConfig{
		Name: "test-workout",
		ItemList: []*config.WorkoutItem{
			{Name: "Jump", Type: config.WorkoutType_Count, Repeat: 2, Target: 10, Break: 5},
			{Name: "Plank", Type: config.WorkoutType_Duration, Repeat: 1, Target: 30, Break: 0},
		},
	}
}

func newTestWCList() []*config.WorkoutConfig {
	return []*config.WorkoutConfig{
		{Name: "A"},
		{Name: "B"},
	}
}

// ==================== initialModel tests ====================

func TestInitialModel_CountItem(t *testing.T) {
	wc := newTestWC()
	m := initialModel(nil, wc)

	assert.Equal(t, WorkoutStatus_Active, m.status, "Count 类型自动激活")
	assert.Equal(t, wc, m.wc)
	assert.Equal(t, 0, m.curItemIdx)
	assert.Equal(t, "Jump", m.curItem.Name)
	assert.Equal(t, 0, m.countDown) // Count 类型不设置 countDown
	assert.Equal(t, 0, m.curRepeat)
}

func TestInitialModel_DurationItem(t *testing.T) {
	wc := &config.WorkoutConfig{
		Name: "timed",
		ItemList: []*config.WorkoutItem{
			{Name: "Plank", Type: config.WorkoutType_Duration, Repeat: 1, Target: 30, Break: 0},
		},
	}
	m := initialModel(nil, wc)
	assert.Equal(t, WorkoutStatus_Inactive, m.status, "Duration 类型初始为 Inactive")
	assert.Equal(t, 30, m.countDown)
}

func TestInitialModel_EmptyItemList(t *testing.T) {
	wc := &config.WorkoutConfig{Name: "empty", ItemList: nil}
	m := initialModel(nil, wc)
	assert.Equal(t, WorkoutStatus_Fin, m.status)
	assert.Nil(t, m.curItem)
}

func TestInitialModel_NilWorkout(t *testing.T) {
	m := initialModel(nil, nil)
	assert.Equal(t, WorkoutStatus_Fin, m.status)
	assert.Nil(t, m.curItem)
}

// ==================== Init tests ====================

func TestInit_ReturnsTickCmd(t *testing.T) {
	wc := newTestWC()
	m := initialModel(nil, wc)
	cmd := m.Init()
	require.NotNil(t, cmd)
}

// ==================== Update tests ====================

func TestUpdate_QuitKeys(t *testing.T) {
	for _, key := range []string{"ctrl+c", "q"} {
		t.Run("key="+key, func(t *testing.T) {
			wc := newTestWC()
			m := initialModel(newTestWCList(), wc)
			result, cmd := m.Update(tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune(key),
			})
			require.NotNil(t, cmd)
			_, ok := cmd().(tea.QuitMsg)
			assert.True(t, ok, "expected QuitMsg for key %s", key)
			_, ok = result.(model)
			assert.True(t, ok)
		})
	}
}

func TestUpdate_SpaceTogglesInactiveForDurationItem(t *testing.T) {
	// Duration 类型: initial = Inactive, Space → Active, Space → Inactive
	wc := &config.WorkoutConfig{
		Name: "timed",
		ItemList: []*config.WorkoutItem{
			{Name: "Hold", Type: config.WorkoutType_Duration, Repeat: 1, Target: 10, Break: 0},
		},
	}
	m := initialModel(newTestWCList(), wc)
	require.Equal(t, WorkoutStatus_Inactive, m.status)

	// Space → Active
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m2 := result.(model)
	assert.Equal(t, WorkoutStatus_Active, m2.status)

	// Space again → Inactive
	result, _ = m2.Update(tea.KeyMsg{Type: tea.KeySpace})
	m3 := result.(model)
	assert.Equal(t, WorkoutStatus_Inactive, m3.status)
}

func TestUpdate_EnterAdvancesCountItem(t *testing.T) {
	wc := &config.WorkoutConfig{
		Name: "count-test",
		ItemList: []*config.WorkoutItem{
			{Name: "Pushup", Type: config.WorkoutType_Count, Repeat: 2, Target: 10, Break: 0},
		},
	}
	m := initialModel(newTestWCList(), wc)
	require.Equal(t, WorkoutStatus_Active, m.status, "Count 类型自动激活")
	require.Equal(t, 0, m.curRepeat)

	// Enter → 完成一次（Repeat=2 需两次）
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := result.(model)
	assert.Equal(t, 1, m2.curRepeat, "第 1 组完成")

	// Enter → 两组完成，列表中只有一个 item 所以结束
	result, _ = m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := result.(model)
	// 完成时 curRepeat 被重置为 0
	assert.Equal(t, 0, m3.curRepeat, "完成时 repeat 归零")
	assert.Equal(t, WorkoutStatus_Fin, m3.status, "所有 item 完成")
	assert.Nil(t, m3.curItem)
}

func TestUpdate_EnterWithBreak(t *testing.T) {
	wc := &config.WorkoutConfig{
		Name: "break-test",
		ItemList: []*config.WorkoutItem{
			{Name: "Squat", Type: config.WorkoutType_Count, Repeat: 1, Target: 10, Break: 5},
		},
	}
	m := initialModel(newTestWCList(), wc)
	require.Equal(t, WorkoutStatus_Active, m.status)

	// Enter → 完成一组，进入休息
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := result.(model)
	assert.True(t, m2.breaking, "进入休息")
	assert.Equal(t, 5, m2.countDown, "休息倒计时 = Break")
}

func TestUpdate_TickDecrementsCountDown(t *testing.T) {
	wc := &config.WorkoutConfig{
		Name: "timed",
		ItemList: []*config.WorkoutItem{
			{Name: "Hold", Type: config.WorkoutType_Duration, Repeat: 1, Target: 10, Break: 0},
		},
	}
	m := initialModel(newTestWCList(), wc)
	require.Equal(t, 10, m.countDown)

	// 激活后 tick 才减少倒计时
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m2 := result.(model)
	require.Equal(t, WorkoutStatus_Active, m2.status)

	result, cmd := m2.Update(tickMsg(time.Now()))
	m3 := result.(model)
	assert.Equal(t, 9, m3.countDown)
	assert.NotNil(t, cmd, "tick 应同时调度下一个 tick")
}

func TestUpdate_TickWhenInactiveDoesNotChangeCountDown(t *testing.T) {
	wc := &config.WorkoutConfig{
		Name: "timed",
		ItemList: []*config.WorkoutItem{
			{Name: "Hold", Type: config.WorkoutType_Duration, Repeat: 1, Target: 10, Break: 0},
		},
	}
	m := initialModel(newTestWCList(), wc)
	require.Equal(t, WorkoutStatus_Inactive, m.status)
	require.Equal(t, 10, m.countDown)

	result, cmd := m.Update(tickMsg(time.Now()))
	m2 := result.(model)
	assert.Equal(t, 10, m2.countDown, "Inactive 时 tick 不应改变倒计时")
	assert.NotNil(t, cmd)
}

func TestUpdate_MouseIncrementsCount(t *testing.T) {
	wc := &config.WorkoutConfig{
		Name: "mouse-test",
		ItemList: []*config.WorkoutItem{
			{Name: "Curl", Type: config.WorkoutType_Count, Repeat: 1, Target: 10, Break: 0},
		},
	}
	m := initialModel(newTestWCList(), wc)
	require.Equal(t, WorkoutStatus_Active, m.status)

	// 鼠标点击 → countDown 从 0 变为 1
	result, _ := m.Update(tea.MouseMsg{Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})
	m2 := result.(model)
	assert.Equal(t, 1, m2.countDown)
	require.Equal(t, 0, m2.curRepeat) // 未按 Enter 前不计入 repeat
}

// ==================== View tests ====================

func TestView_FinStatus(t *testing.T) {
	wc := &config.WorkoutConfig{Name: "done", ItemList: nil}
	m := initialModel(newTestWCList(), wc)
	require.Equal(t, WorkoutStatus_Fin, m.status)

	view := m.View()
	// FIN 状态渲染 ASCII art，验证输出不为空且包含工作列表
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "A")
	assert.Contains(t, view, "B")
}

func TestView_ActiveItem(t *testing.T) {
	wc := newTestWC()
	m := initialModel(newTestWCList(), wc)
	require.Equal(t, WorkoutStatus_Active, m.status)

	view := m.View()
	assert.Contains(t, view, "Jump")
	assert.NotEmpty(t, view)
}

func TestView_DisplaysWorkoutList(t *testing.T) {
	wc := newTestWC()
	m := initialModel(newTestWCList(), wc)

	view := m.View()
	// 应包含两个 workconfig 名称
	assert.Contains(t, view, "A")
	assert.Contains(t, view, "B")
}
