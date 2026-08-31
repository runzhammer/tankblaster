//go:build android || ios
// +build android ios

package tankblaster

import "image"

type shopTouchScrollState struct {
	active     bool
	startY     int
	start      int
	dragged    bool
	blockClick bool
}

func (s *GameScene) handleMobileShopListScroll(count int) bool {
	if count <= 0 {
		s.shopTouchScroll = shopTouchScrollState{}
		return false
	}
	if primaryPointerJustPressed() {
		x, y := primaryPointerPosition()
		if image.Pt(x, y).In(s.shopListRect()) {
			s.shopTouchScroll = shopTouchScrollState{
				active: true,
				startY: y,
				start:  s.shopSelectedIndex,
			}
		}
	}
	if s.shopTouchScroll.blockClick {
		s.shopTouchScroll.blockClick = false
		return true
	}
	if !s.shopTouchScroll.active {
		return false
	}
	if primaryPointerPressed() {
		_, y := primaryPointerPosition()
		delta := s.shopTouchScroll.startY - y
		rowH := maxInt(1, s.shopListRowHeight())
		if delta <= -rowH/3 || delta >= rowH/3 {
			s.shopTouchScroll.dragged = true
			s.setShopSelectedIndex(maxInt(0, minInt(count-1, s.shopTouchScroll.start+delta/rowH)))
		}
		return s.shopTouchScroll.dragged
	}
	if primaryPointerJustReleased() {
		dragged := s.shopTouchScroll.dragged
		s.shopTouchScroll = shopTouchScrollState{blockClick: dragged}
		return dragged
	}
	s.shopTouchScroll = shopTouchScrollState{}
	return false
}
