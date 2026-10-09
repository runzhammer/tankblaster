package tankblaster

import (
	"hash/fnv"
	"math"
	"math/rand"
	"strconv"

	"github.com/runzhammer/tankblaster/pkg/core"
	"github.com/runzhammer/tankblaster/pkg/protocol"
)

func (s *GameScene) consumeOnlineGameCommands() {
	if s.g.online == nil || s.g.online.client == nil {
		return
	}
	for {
		select {
		case env := <-s.g.online.client.recv:
			s.handleOnlineGameEnvelope(env)
		case err := <-s.g.online.client.errs:
			if err != nil {
				// Keep the offline scene alive; network errors should not crash rendering.
			}
		default:
			return
		}
	}
}

func (s *GameScene) handleOnlineGameEnvelope(env protocol.Envelope) {
	if !s.acceptOnlineGameEnvelope(env) {
		return
	}
	switch env.Type {
	case protocol.TypeHelloAck:
		s.requestOnlineCatchUp()
	case protocol.TypeOnlineGameCommand:
		cmd, err := protocol.Decode[protocol.OnlineGameCommand](env)
		if err == nil {
			s.applyOnlineGameCommand(cmd)
		}
	case protocol.TypeStateUpdate, protocol.TypeGameStart, protocol.TypeTurnStart, protocol.TypeGameOver:
		msg, err := protocol.Decode[protocol.StateUpdate](env)
		if err == nil {
			s.g.online.state = msg.State
		}
	}
}

func (s *GameScene) acceptOnlineGameEnvelope(env protocol.Envelope) bool {
	if s.g.online == nil {
		return false
	}
	if env.MatchID != "" && env.MatchID != s.g.online.state.MatchID {
		return false
	}
	if env.Sequence <= 0 {
		return true
	}
	if s.g.online.lastServerSequence == 0 {
		s.g.online.lastServerSequence = env.Sequence
		s.g.online.awaitingCatchUp = false
		return true
	}
	if env.Sequence <= s.g.online.lastServerSequence {
		return false
	}
	if env.Sequence != s.g.online.lastServerSequence+1 {
		s.requestOnlineCatchUp()
		return false
	}
	s.g.online.lastServerSequence = env.Sequence
	s.g.online.awaitingCatchUp = false
	return true
}

func (s *GameScene) requestOnlineCatchUp() {
	if s.g.online == nil || s.g.online.client == nil || s.g.online.awaitingCatchUp || s.g.online.state.MatchID == "" {
		return
	}
	s.g.online.awaitingCatchUp = true
	s.g.online.client.Send(protocol.TypeReconnect, protocol.Reconnect{
		MatchID:       s.g.online.state.MatchID,
		AfterSequence: s.g.online.lastServerSequence,
	})
}

func (s *GameScene) sendOnlineGameCommand(cmd protocol.OnlineGameCommand) {
	if s.g.online == nil || s.g.online.client == nil {
		return
	}
	s.g.online.commandSequence++
	cmd.MatchID = s.g.online.state.MatchID
	if cmd.PlayerID == "" && cmd.PlayerIndex >= 0 && cmd.PlayerIndex < len(s.g.online.state.Players) {
		cmd.PlayerID = s.g.online.state.Players[cmd.PlayerIndex].ID
	}
	if cmd.PlayerID == "" {
		cmd.PlayerID = s.g.online.playerID
	}
	cmd.CommandSequence = s.g.online.commandSequence
	s.attachOnlineSharedState(&cmd)
	s.g.online.client.Send(protocol.TypeOnlineGameCommand, cmd)
}

func (s *GameScene) applyOnlineGameCommand(cmd protocol.OnlineGameCommand) {
	if s.g.online == nil || cmd.MatchID != s.g.online.state.MatchID {
		return
	}
	s.applyOnlineSharedState(cmd)
	switch cmd.Kind {
	case "aim":
		s.applyOnlineAimCommand(cmd)
	case "fire":
		s.applyOnlineFireCommand(cmd)
	case "xm_v12":
		s.applyOnlineXMV12Command(cmd)
	case "turn":
		s.applyOnlineTurnCommand(cmd)
	case "shop_hover":
		s.applyOnlineShopHoverCommand(cmd)
	case "shop_mode":
		s.applyOnlineShopModeCommand(cmd)
	case "shop_select":
		s.applyOnlineShopSelectCommand(cmd)
	case "shop_buy":
		s.applyOnlineShopBuyCommand(cmd)
	case "shop_back":
		s.applyOnlineShopBackCommand(cmd)
	case "shop_continue":
		s.applyOnlineShopContinueCommand(cmd)
	}
}

func (s *GameScene) applyOnlineShopHoverCommand(cmd protocol.OnlineGameCommand) {
	if !s.onlineShopCommandAllowed(cmd) {
		return
	}
	if s.onlineCommandIsLocalControlled(cmd) {
		return
	}
	s.shopHoverClass = maxInt(0, minInt(2, cmd.ShopHoverClass))
}

func (s *GameScene) applyOnlineShopModeCommand(cmd protocol.OnlineGameCommand) {
	if !s.onlineShopCommandAllowed(cmd) {
		return
	}
	if s.onlineCommandIsLocalControlled(cmd) {
		return
	}
	s.withRemoteOnlineCommand(func() {
		mode := shopMode(cmd.ShopMode)
		if mode != shopModeClassA && mode != shopModeClassB {
			return
		}
		s.setShopMode(mode)
	})
}

func (s *GameScene) applyOnlineShopSelectCommand(cmd protocol.OnlineGameCommand) {
	if !s.onlineShopCommandAllowed(cmd) {
		return
	}
	if s.onlineCommandIsLocalControlled(cmd) {
		return
	}
	s.withRemoteOnlineCommand(func() {
		s.setShopSelectedIndex(cmd.ShopListIndex)
	})
}

func (s *GameScene) applyOnlineShopBuyCommand(cmd protocol.OnlineGameCommand) {
	if !s.onlineShopCommandAllowed(cmd) {
		return
	}
	if s.onlineCommandIsLocalControlled(cmd) {
		return
	}
	if cmd.Economy != nil {
		return
	}
	s.withRemoteOnlineCommand(func() {
		s.buySelectedShopItem()
	})
}

func (s *GameScene) applyOnlineShopBackCommand(cmd protocol.OnlineGameCommand) {
	if !s.onlineShopCommandAllowed(cmd) {
		return
	}
	if s.onlineCommandIsLocalControlled(cmd) {
		return
	}
	s.withRemoteOnlineCommand(func() {
		s.shopMode = shopModeEntry
		s.shopHoverClass = 0
	})
}

func (s *GameScene) applyOnlineShopContinueCommand(cmd protocol.OnlineGameCommand) {
	if !s.onlineShopCommandAllowed(cmd) {
		return
	}
	if s.onlineCommandIsLocalControlled(cmd) {
		return
	}
	s.withRemoteOnlineCommand(func() {
		s.continueShop()
	})
}

func (s *GameScene) onlineShopCommandAllowed(cmd protocol.OnlineGameCommand) bool {
	if s.g.online == nil || cmd.MatchID != s.g.online.state.MatchID || s.phase != phaseShop {
		return false
	}
	if s.onlineCommandIsLocalControlled(cmd) {
		s.shopAwaitingOnline = false
	}
	return s.onlinePlayerIndexForCommand(cmd) == s.currentShopPlayerIndex()
}

func (s *GameScene) withRemoteOnlineCommand(apply func()) {
	s.g.online.applyingRemoteCmd = true
	defer func() {
		s.g.online.applyingRemoteCmd = false
	}()
	apply()
}

func (s *GameScene) applyOnlineFireCommand(cmd protocol.OnlineGameCommand) {
	if !s.onlineCommandIsCurrentTurn(cmd) {
		return
	}
	if !s.onlineCommandSequenceIsFresh(cmd) {
		return
	}
	s.g.online.applyingRemoteCmd = true
	defer func() {
		s.g.online.applyingRemoteCmd = false
	}()
	playerIndex := s.onlinePlayerIndexForCommand(cmd)
	if playerIndex < 0 || playerIndex >= len(s.tanks) {
		return
	}
	tank := s.tanks[playerIndex]
	if tank == nil || tank.cannon == nil {
		return
	}
	s.activePlayerIndex = playerIndex
	if cmd.WeaponSlot >= 0 && cmd.WeaponSlot < s.weaponSlotCount() {
		tank.selectedWeapon = cmd.WeaponSlot
	}
	tank.shotStrength = maxInt(0, minInt(cmd.ShotStrength, s.maxShotStrength()))
	tank.cannon.Rot = cmd.CannonRotation
	s.cameraX = s.clampOnlineCameraX(cmd.CameraX)
	s.withDeterministicOnlineRNG(cmd, func() {
		s.fireActiveWeapon()
	})
}

func (s *GameScene) applyOnlineAimCommand(cmd protocol.OnlineGameCommand) {
	if !s.onlineCommandIsCurrentTurn(cmd) {
		return
	}
	if s.onlineCommandIsLocalControlled(cmd) {
		return
	}
	if !s.onlineCommandSequenceIsFresh(cmd) {
		return
	}
	s.g.online.applyingRemoteCmd = true
	defer func() {
		s.g.online.applyingRemoteCmd = false
	}()
	tank := s.tankForOnlineCommand(cmd)
	if tank == nil || tank.cannon == nil {
		return
	}
	if s.activePlayerIndex >= 0 && tank.playerIndex != s.activePlayerIndex {
		return
	}
	if s.activePlayerIndex < 0 {
		s.activePlayerIndex = tank.playerIndex
	}
	if cmd.WeaponSlot >= 0 && cmd.WeaponSlot < s.weaponSlotCount() {
		tank.selectedWeapon = cmd.WeaponSlot
	}
	tank.shotStrength = maxInt(s.minShotStrength(), minInt(cmd.ShotStrength, s.maxShotStrength()))
	tank.cannon.Rot = cmd.CannonRotation
	s.clampCannonRotationToTank(tank.cannon, tank.body)
	s.updateXMV12FacingFromCannon(tank)
	s.cameraX = s.clampOnlineCameraX(cmd.CameraX)
}

func (s *GameScene) applyOnlineXMV12Command(cmd protocol.OnlineGameCommand) {
	if !s.onlineCommandIsCurrentTurn(cmd) {
		return
	}
	s.g.online.applyingRemoteCmd = true
	defer func() {
		s.g.online.applyingRemoteCmd = false
	}()
	tank := s.tankForOnlineCommand(cmd)
	if tank == nil {
		return
	}
	s.activePlayerIndex = tank.playerIndex
	s.xmV12DriveMode = cmd.XMV12Mode
	s.xmV12DriveDirection = cmd.XMV12Direction
	if cmd.XMV12MotorOff {
		s.playEventSound(soundEventXMV12MotorOff)
		s.stopXMV12LoopSounds()
		s.xmV12DriveMode = true
		s.xmV12DriveDirection = 0
		s.xmV12EngineOffDelay = xmV12EngineOffDelayFrames
		return
	}
	if cmd.XMV12Mode {
		s.xmV12EngineOffDelay = 0
		s.playEventSound(soundEventXMV12Ignition)
	}
}

func (s *GameScene) applyOnlineTurnCommand(cmd protocol.OnlineGameCommand) {
	if cmd.TurnSequence <= s.g.online.turnSequence {
		return
	}
	if s.onlineTurnAdvanceInProgress() {
		return
	}
	s.g.online.applyingRemoteCmd = true
	defer func() {
		s.g.online.applyingRemoteCmd = false
	}()
	playerIndex := cmd.PlayerIndex
	if playerIndex < 0 || playerIndex >= len(s.tanks) {
		return
	}
	s.g.online.turnSequence = cmd.TurnSequence
	s.activePlayerIndex = playerIndex
	s.applyOnlineCloudStates(cmd.Clouds)
	s.resetComputerTurnPlans()
	s.clampActiveShotStrength()
	s.cameraGoal = s.cameraTargetForTank(s.activePlayerIndex)
	s.cameraGoalY = 0
	s.cameraX = s.clampOnlineCameraX(cmd.CameraX)
	s.cameraY = 0
}

func (s *GameScene) syncOnlineAim(tank *battleTank) {
	if s.g.online == nil || s.g.online.applyingRemoteCmd || tank == nil || tank.cannon == nil || !s.onlineCanControlPlayer(tank.playerIndex) {
		return
	}
	state := onlineAimState{
		valid:          true,
		playerIndex:    tank.playerIndex,
		turnSequence:   s.g.online.turnSequence,
		weaponSlot:     tank.selectedWeapon,
		shotStrength:   tank.shotStrength,
		cannonRotation: tank.cannon.Rot,
		cameraX:        s.cameraX,
	}
	if s.onlineAimStateUnchanged(state) {
		return
	}
	s.g.online.lastAim = state
	s.sendOnlineGameCommand(protocol.OnlineGameCommand{
		Kind:           "aim",
		PlayerIndex:    tank.playerIndex,
		TurnSequence:   s.g.online.turnSequence,
		WeaponSlot:     tank.selectedWeapon,
		ShotStrength:   tank.shotStrength,
		CannonRotation: tank.cannon.Rot,
		CameraX:        s.cameraX,
	})
}

func (s *GameScene) requestXMV12Start(tank *battleTank) {
	if tank == nil {
		return
	}
	s.playEventSound(soundEventXMV12Ignition)
	s.xmV12DriveMode = true
	s.xmV12DriveDirection = 0
	s.syncOnlineXMV12(tank, false)
}

func (s *GameScene) requestXMV12Direction(tank *battleTank, direction int) {
	if tank == nil {
		return
	}
	s.xmV12DriveDirection = direction
	s.syncOnlineXMV12(tank, false)
}

func (s *GameScene) requestXMV12MotorOff(tank *battleTank) {
	if tank == nil {
		return
	}
	s.playEventSound(soundEventXMV12MotorOff)
	s.stopXMV12LoopSounds()
	s.xmV12DriveDirection = 0
	s.xmV12EngineOffDelay = xmV12EngineOffDelayFrames
	s.syncOnlineXMV12(tank, true)
}

func (s *GameScene) syncOnlineXMV12(tank *battleTank, motorOff bool) {
	if s.g.online == nil || s.g.online.applyingRemoteCmd || tank == nil || !s.onlineCanControlPlayer(tank.playerIndex) {
		return
	}
	s.sendOnlineGameCommand(protocol.OnlineGameCommand{
		Kind:           "xm_v12",
		PlayerIndex:    tank.playerIndex,
		TurnSequence:   s.g.online.turnSequence,
		CameraX:        s.cameraX,
		XMV12Mode:      s.xmV12DriveMode,
		XMV12Direction: s.xmV12DriveDirection,
		XMV12MotorOff:  motorOff,
	})
}

func (s *GameScene) syncOnlineTurn(previousPlayerIndex int) {
	if s.g.online == nil || s.g.online.applyingRemoteCmd {
		return
	}
	s.g.online.turnSequence++
	if !s.onlineCanControlPlayer(previousPlayerIndex) {
		return
	}
	s.sendOnlineGameCommand(protocol.OnlineGameCommand{
		Kind:         "turn",
		PlayerID:     s.g.online.state.Players[previousPlayerIndex].ID,
		PlayerIndex:  s.activePlayerIndex,
		TurnSequence: s.g.online.turnSequence,
		CameraX:      s.cameraX,
		Clouds:       s.onlineCloudStates(),
	})
}

func (s *GameScene) syncOnlineShopHover(hoverClass int) {
	if s.g.online == nil || s.g.online.applyingRemoteCmd || !s.onlineCanControlPlayer(s.currentShopPlayerIndex()) {
		return
	}
	playerIndex := s.currentShopPlayerIndex()
	s.shopHoverClass = maxInt(0, minInt(2, hoverClass))
	s.sendOnlineGameCommand(protocol.OnlineGameCommand{
		Kind:           "shop_hover",
		PlayerIndex:    playerIndex,
		ShopHoverClass: maxInt(0, minInt(2, hoverClass)),
	})
}

func (s *GameScene) syncOnlineShopMode(mode shopMode) {
	if !s.canSendOnlineShopCommand() {
		return
	}
	playerIndex := s.currentShopPlayerIndex()
	s.setShopMode(mode)
	s.sendOnlineGameCommand(protocol.OnlineGameCommand{
		Kind:        "shop_mode",
		PlayerIndex: playerIndex,
		ShopMode:    int(mode),
	})
}

func (s *GameScene) syncOnlineShopSelect(index int) {
	if !s.canSendOnlineShopCommand() {
		return
	}
	playerIndex := s.currentShopPlayerIndex()
	s.setShopSelectedIndex(index)
	s.sendOnlineGameCommand(protocol.OnlineGameCommand{
		Kind:          "shop_select",
		PlayerIndex:   playerIndex,
		ShopListIndex: index,
	})
}

func (s *GameScene) syncOnlineShopBuy() {
	if !s.canSendOnlineShopCommand() {
		return
	}
	playerIndex := s.currentShopPlayerIndex()
	s.buySelectedShopItem()
	s.sendOnlineGameCommand(protocol.OnlineGameCommand{
		Kind:        "shop_buy",
		PlayerIndex: playerIndex,
	})
}

func (s *GameScene) syncOnlineShopBack() {
	if !s.canSendOnlineShopCommand() {
		return
	}
	playerIndex := s.currentShopPlayerIndex()
	s.shopMode = shopModeEntry
	s.shopHoverClass = 0
	s.sendOnlineGameCommand(protocol.OnlineGameCommand{
		Kind:        "shop_back",
		PlayerIndex: playerIndex,
	})
}

func (s *GameScene) syncOnlineShopContinue() {
	if !s.canSendOnlineShopCommand() {
		return
	}
	playerIndex := s.currentShopPlayerIndex()
	s.continueShop()
	s.sendOnlineGameCommand(protocol.OnlineGameCommand{
		Kind:        "shop_continue",
		PlayerIndex: playerIndex,
	})
}

func (s *GameScene) canSendOnlineShopCommand() bool {
	return s.g.online != nil &&
		!s.g.online.applyingRemoteCmd &&
		s.phase == phaseShop &&
		s.onlineCanControlPlayer(s.currentShopPlayerIndex())
}

func (s *GameScene) onlineAimStateUnchanged(next onlineAimState) bool {
	prev := s.g.online.lastAim
	if !prev.valid {
		return false
	}
	const epsilon = 0.000001
	return prev.playerIndex == next.playerIndex &&
		prev.turnSequence == next.turnSequence &&
		prev.weaponSlot == next.weaponSlot &&
		prev.shotStrength == next.shotStrength &&
		math.Abs(prev.cannonRotation-next.cannonRotation) <= epsilon &&
		math.Abs(prev.cameraX-next.cameraX) <= epsilon
}

func (s *GameScene) onlineCommandSequenceIsFresh(cmd protocol.OnlineGameCommand) bool {
	if cmd.CommandSequence <= 0 {
		return true
	}
	if s.g.online.receivedCommandSequences == nil {
		s.g.online.receivedCommandSequences = make(map[string]int)
	}
	key := cmd.PlayerID + ":" + cmd.Kind
	if cmd.CommandSequence <= s.g.online.receivedCommandSequences[key] {
		return false
	}
	s.g.online.receivedCommandSequences[key] = cmd.CommandSequence
	return true
}

func (s *GameScene) onlineCloudStates() []protocol.OnlineCloudState {
	if len(s.clouds) == 0 {
		return nil
	}
	states := make([]protocol.OnlineCloudState, 0, len(s.clouds))
	for _, cloud := range s.clouds {
		if cloud == nil || cloud.sprite == nil || cloud.sprite.Pos == nil {
			continue
		}
		states = append(states, protocol.OnlineCloudState{
			Kind:       int(cloud.kind),
			X:          cloud.sprite.Pos.X,
			Y:          cloud.sprite.Pos.Y,
			Speed:      cloud.speed,
			Aggression: cloud.aggression,
		})
	}
	return states
}

func (s *GameScene) applyOnlineCloudStates(states []protocol.OnlineCloudState) {
	if len(states) == 0 || len(s.clouds) == 0 {
		return
	}
	count := minInt(len(states), len(s.clouds))
	for i := 0; i < count; i++ {
		cloud := s.clouds[i]
		if cloud == nil || cloud.sprite == nil || cloud.sprite.Pos == nil {
			continue
		}
		state := states[i]
		if int(cloud.kind) != state.Kind {
			continue
		}
		cloud.sprite.Pos.X = state.X
		cloud.sprite.Pos.Y = state.Y
		if state.Speed > 0 {
			cloud.speed = state.Speed
		}
		cloud.aggression = state.Aggression
	}
}

func (s *GameScene) onlinePalmStates() []protocol.OnlinePalmState {
	if len(s.palms) == 0 {
		return nil
	}
	states := make([]protocol.OnlinePalmState, 0, len(s.palms))
	for _, palm := range s.palms {
		if palm == nil {
			continue
		}
		states = append(states, protocol.OnlinePalmState{
			State:                int(palm.state),
			Age:                  palm.age,
			EyeAge:               palm.eyeAge,
			EyesOn:               palm.eyesOn,
			ScreamAge:            palm.screamAge,
			Screaming:            palm.screaming,
			GrinAge:              palm.grinAge,
			Grinning:             palm.grinning,
			GrinHideAt:           palm.grinHideAt,
			Aggression:           palm.aggression,
			InitialAggression:    palm.initialAggression,
			AggressionMultiplier: palm.aggressionMultiplier,
		})
	}
	return states
}

func (s *GameScene) applyOnlinePalmStates(states []protocol.OnlinePalmState) {
	if len(states) == 0 || len(s.palms) == 0 {
		return
	}
	count := minInt(len(states), len(s.palms))
	for i := 0; i < count; i++ {
		palm := s.palms[i]
		if palm == nil {
			continue
		}
		state := states[i]
		palm.state = palmState(maxInt(int(palmStateAlive), minInt(int(palmStateCrumbling), state.State)))
		palm.age = state.Age
		palm.eyeAge = state.EyeAge
		palm.eyesOn = state.EyesOn
		palm.screamAge = state.ScreamAge
		palm.screaming = state.Screaming
		palm.grinAge = state.GrinAge
		palm.grinning = state.Grinning
		palm.grinHideAt = state.GrinHideAt
		palm.aggression = state.Aggression
		palm.initialAggression = state.InitialAggression
		if state.AggressionMultiplier > 0 {
			palm.aggressionMultiplier = state.AggressionMultiplier
		}
	}
}

func (s *GameScene) onlineEconomyState() *protocol.OnlineEconomy {
	if len(s.scores) == 0 && len(s.roundScores) == 0 && len(s.credits) == 0 && len(s.inventories) == 0 {
		return nil
	}
	economy := &protocol.OnlineEconomy{
		Scores:       append([]int(nil), s.scores...),
		RoundScores:  append([]int(nil), s.roundScores...),
		Credits:      append([]int(nil), s.credits...),
		ShopOrder:    append([]int(nil), s.shopPlayerOrder...),
		ShopCursor:   s.shopPlayerCursor,
		ShopMode:     int(s.shopMode),
		ShopHover:    s.shopHoverClass,
		ShopSelected: s.shopSelectedIndex,
	}
	if len(s.inventories) > 0 {
		economy.Inventories = make([]protocol.OnlineInventoryState, 0, len(s.inventories))
		for _, inventory := range s.inventories {
			economy.Inventories = append(economy.Inventories, protocol.OnlineInventoryState{
				ClassA:            append([]int(nil), inventory.classA...),
				ClassB:            append([]int(nil), inventory.classB...),
				MFSBoosterCharges: inventory.mfsBoosterCharges,
				EnergyShield:      inventory.energyShield,
				HasXMV12:          inventory.hasXMV12,
				Diesel:            inventory.diesel,
			})
		}
	}
	return economy
}

func (s *GameScene) applyOnlineEconomyState(economy *protocol.OnlineEconomy) {
	if economy == nil {
		return
	}
	if len(economy.Scores) > 0 {
		s.scores = append([]int(nil), economy.Scores...)
	}
	if len(economy.RoundScores) > 0 {
		s.roundScores = append([]int(nil), economy.RoundScores...)
	}
	if len(economy.Credits) > 0 {
		s.credits = append([]int(nil), economy.Credits...)
	}
	if len(economy.Inventories) > 0 {
		s.inventories = make([]shopInventory, len(economy.Inventories))
		for i, inventory := range economy.Inventories {
			s.inventories[i] = shopInventory{
				classA:            append([]int(nil), inventory.ClassA...),
				classB:            append([]int(nil), inventory.ClassB...),
				mfsBoosterCharges: inventory.MFSBoosterCharges,
				energyShield:      inventory.EnergyShield,
				hasXMV12:          inventory.HasXMV12,
				diesel:            inventory.Diesel,
			}
			s.ensureInventory(i)
		}
	}
	if len(economy.ShopOrder) > 0 {
		s.shopPlayerOrder = append([]int(nil), economy.ShopOrder...)
		s.shopPlayerCursor = maxInt(0, minInt(economy.ShopCursor, len(s.shopPlayerOrder)-1))
	}
	s.shopMode = shopMode(maxInt(int(shopModeEntry), minInt(int(shopModeClassB), economy.ShopMode)))
	s.shopHoverClass = maxInt(0, minInt(2, economy.ShopHover))
	s.shopSelectedIndex = maxInt(0, economy.ShopSelected)
}

func (s *GameScene) attachOnlineSharedState(cmd *protocol.OnlineGameCommand) {
	if cmd == nil {
		return
	}
	if len(cmd.Clouds) == 0 {
		cmd.Clouds = s.onlineCloudStates()
	}
	if len(cmd.Palms) == 0 {
		cmd.Palms = s.onlinePalmStates()
	}
	if cmd.Economy == nil {
		cmd.Economy = s.onlineEconomyState()
	}
}

func (s *GameScene) applyOnlineSharedState(cmd protocol.OnlineGameCommand) {
	s.applyOnlineCloudStates(cmd.Clouds)
	s.applyOnlinePalmStates(cmd.Palms)
	s.applyOnlineEconomyState(cmd.Economy)
}

func (s *GameScene) onlineCommandIsLocalControlled(cmd protocol.OnlineGameCommand) bool {
	return s.onlineCanControlPlayer(s.onlinePlayerIndexForCommand(cmd))
}

func (s *GameScene) onlineCommandIsCurrentTurn(cmd protocol.OnlineGameCommand) bool {
	return s.g.online == nil || cmd.TurnSequence == s.g.online.turnSequence
}

func (s *GameScene) onlineTurnAdvanceInProgress() bool {
	return s.projectilesActive() || s.turnAdvanceDelay > 0 || s.turnAdvanceBlocked() || s.xmV12DriveMode
}

func (s *GameScene) clampOnlineCameraX(cameraX float64) float64 {
	maxCameraX := s.worldWidth - core.Config().Screen.Width
	if maxCameraX < 0 {
		maxCameraX = 0
	}
	return maxFloat64(0, minFloat(cameraX, maxCameraX))
}

func (s *GameScene) onlineComputerDecisionSeed(tank *battleTank) int64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(strconv.FormatInt(s.g.online.state.Seed, 10)))
	_, _ = hasher.Write([]byte(":"))
	_, _ = hasher.Write([]byte(strconv.Itoa(s.roundNumber)))
	_, _ = hasher.Write([]byte(":"))
	_, _ = hasher.Write([]byte(strconv.Itoa(s.activePlayerIndex)))
	_, _ = hasher.Write([]byte(":"))
	_, _ = hasher.Write([]byte(strconv.Itoa(int(s.time / 30))))
	if tank != nil {
		_, _ = hasher.Write([]byte(":"))
		_, _ = hasher.Write([]byte(tank.player.Name))
	}
	return int64(hasher.Sum64())
}

func (s *GameScene) withDeterministicOnlineRNG(cmd protocol.OnlineGameCommand, apply func()) {
	if s.g.online == nil {
		apply()
		return
	}
	previous := s.rng
	s.rng = rand.New(rand.NewSource(s.onlineCommandSeed(cmd)))
	defer func() {
		s.rng = previous
	}()
	apply()
}

func (s *GameScene) onlineCommandSeed(cmd protocol.OnlineGameCommand) int64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(strconv.FormatInt(s.g.online.state.Seed, 10)))
	_, _ = hasher.Write([]byte(":"))
	_, _ = hasher.Write([]byte(cmd.MatchID))
	_, _ = hasher.Write([]byte(":"))
	_, _ = hasher.Write([]byte(cmd.PlayerID))
	_, _ = hasher.Write([]byte(":"))
	_, _ = hasher.Write([]byte(cmd.Kind))
	_, _ = hasher.Write([]byte(":"))
	_, _ = hasher.Write([]byte(strconv.Itoa(cmd.PlayerIndex)))
	_, _ = hasher.Write([]byte(":"))
	_, _ = hasher.Write([]byte(strconv.Itoa(cmd.TurnSequence)))
	_, _ = hasher.Write([]byte(":"))
	_, _ = hasher.Write([]byte(strconv.Itoa(cmd.CommandSequence)))
	_, _ = hasher.Write([]byte(":"))
	_, _ = hasher.Write([]byte(strconv.Itoa(cmd.WeaponSlot)))
	_, _ = hasher.Write([]byte(":"))
	_, _ = hasher.Write([]byte(strconv.Itoa(cmd.ShotStrength)))
	return int64(hasher.Sum64())
}

func (s *GameScene) zeroPowerAnimationRNG(tank *battleTank) (*rand.Rand, int64) {
	if s.g == nil || s.g.online == nil {
		return s.rng, 0
	}
	seed := s.onlineZeroPowerSeed(tank)
	return rand.New(rand.NewSource(seed)), seed
}

func (s *GameScene) onlineZeroPowerSeed(tank *battleTank) int64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(strconv.FormatInt(s.g.online.state.Seed, 10)))
	_, _ = hasher.Write([]byte(":zero_power:"))
	_, _ = hasher.Write([]byte(s.g.online.state.MatchID))
	_, _ = hasher.Write([]byte(":"))
	_, _ = hasher.Write([]byte(strconv.Itoa(s.roundNumber)))
	_, _ = hasher.Write([]byte(":"))
	_, _ = hasher.Write([]byte(strconv.Itoa(s.g.online.turnSequence)))
	_, _ = hasher.Write([]byte(":"))
	if tank != nil {
		_, _ = hasher.Write([]byte(strconv.Itoa(tank.playerIndex)))
	}
	_, _ = hasher.Write([]byte(":"))
	if s.lastDamageSource != nil {
		_, _ = hasher.Write([]byte(strconv.Itoa(s.lastDamageSource.playerIndex)))
	}
	return int64(hasher.Sum64())
}

func (s *GameScene) reportOnlineMatchComplete() {
	if s.g.online == nil || s.g.online.client == nil || s.g.online.matchResultSent {
		return
	}
	winnerIndex, ok := s.onlineWinnerByScore()
	if !ok || winnerIndex < 0 || winnerIndex >= len(s.g.online.state.Players) {
		return
	}
	if s.g.online.state.Players[winnerIndex].ID != s.g.online.playerID {
		return
	}
	scores := make(map[string]int, len(s.g.online.state.Players))
	for index, player := range s.g.online.state.Players {
		scores[player.ID] = s.scoreForPlayer(index)
	}
	s.g.online.matchResultSent = true
	s.g.online.client.Send(protocol.TypeMatchComplete, protocol.MatchComplete{
		MatchID: s.g.online.state.MatchID,
		Scores:  scores,
	})
}

func (s *GameScene) onlineWinnerByScore() (int, bool) {
	if s.g.online == nil || len(s.g.online.state.Players) == 0 {
		return -1, false
	}
	winnerIndex := -1
	winnerScore := 0
	tied := false
	for index := range s.g.online.state.Players {
		score := s.scoreForPlayer(index)
		if winnerIndex < 0 || score > winnerScore {
			winnerIndex = index
			winnerScore = score
			tied = false
			continue
		}
		if score == winnerScore {
			tied = true
		}
	}
	return winnerIndex, winnerIndex >= 0 && !tied
}

func (s *GameScene) tankForOnlineCommand(cmd protocol.OnlineGameCommand) *battleTank {
	playerIndex := s.onlinePlayerIndexForCommand(cmd)
	if playerIndex < 0 || playerIndex >= len(s.tanks) {
		return nil
	}
	return s.tanks[playerIndex]
}

func maxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func (s *GameScene) onlinePlayerIndexForCommand(cmd protocol.OnlineGameCommand) int {
	if cmd.PlayerIndex >= 0 &&
		cmd.PlayerIndex < len(s.g.online.state.Players) &&
		s.g.online.state.Players[cmd.PlayerIndex].ID == cmd.PlayerID {
		return cmd.PlayerIndex
	}
	for index, player := range s.g.online.state.Players {
		if player.ID == cmd.PlayerID {
			return index
		}
	}
	return -1
}
