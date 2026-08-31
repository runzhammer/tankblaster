package tankblaster

type languageID uint8

const (
	languageGerman languageID = iota
	languageEnglish
)

type localizedStrings struct {
	PlayerSelectionTitle                 string
	PlayerSelectionRounds                string
	PlayerSelectionSlotNone              string
	PlayerSelectionSlotHuman             string
	PlayerSelectionSlotComputer          string
	PlayerSelectionHelpHint              string
	PlayerSelectionOptionsButton         string
	PlayerSelectionOnlineButton          string
	PlayerSelectionStartButton           string
	PlayerSelectionMinimumPlayers        string
	PlayerSelectionHelpTitle             string
	PlayerSelectionHelpSlotNone          string
	PlayerSelectionHelpHuman             string
	PlayerSelectionHelpPlayerName        string
	PlayerSelectionHelpPlayerType        string
	PlayerSelectionHelpColor             string
	PlayerSelectionHelpName              string
	PlayerSelectionHelpComputerType      string
	PlayerSelectionHelpKeys              string
	PlayerSelectionHelpKeyIncreaseRounds string
	PlayerSelectionHelpKeyDecreaseRounds string
	PlayerSelectionHelpKeyOptions        string
	PlayerSelectionHelpKeyStart          string
	OptionsTitle                         string
	OptionsProjectileReentry             string
	OptionsProjectileReentryOff          string
	OptionsProjectileReentryAlways       string
	OptionsProjectileReentryRandom       string
	OptionsCloudAggression               string
	OptionsPalmCount                     string
	OptionsRoundStart                    string
	OptionsQuickRoundStart               string
	OptionsLanguageButton                string
	LanguageTitle                        string
	LanguageGerman                       string
	LanguageEnglish                      string
	OnlineTitle                          string
	OnlineDisplayName                    string
	OnlineQuickMatch                     string
	OnlineCreatePublicSession            string
	OnlineCreatePrivateSession           string
	OnlineOpenSessions                   string
	OnlineJoinSession                    string
	OnlineLeaderboard                    string
	OnlineReady                          string
	OnlineLeave                          string
	OnlineBack                           string
	OnlineJoinPrompt                     string
	OnlineConnecting                     string
	OnlineConnected                      string
	OnlineQueued                         string
	OnlineSessionCreated                 string
	OnlineSessionJoined                  string
	OnlineInviteLink                     string
	OnlineCopiedCode                     string
	OnlineCopiedInviteLink               string
	OnlineClipboardUnavailable           string
	OnlineNoSessions                     string
	OnlineNoLeaderboard                  string
	OnlineMatchStarted                   string
	GameDefaultPlayerName                string
	GameHUDStrength                      string
	GameHUDAngle                         string
	GameHUDFire                          string
	GameHUDIgnition                      string
	GameHUDWind                          string
	GameHUDPower                         string
	GameScoreRound                       string
	GameScoreOf                          string
	GameScorePlayer                      string
	GameScoreSuccess                     string
	GameScoreStatus                      string
	GameStatusActive                     string
	GameStatusOut                        string
	GameStatusEliminated                 string
	GamePause                            string
	GameHelpTitle                        string
	GameHelpKeys                         string
	GameHelpFire                         string
	GameHelpNextWeapon                   string
	GameHelpRotateClockwise              string
	GameHelpRotateCounterClockwise       string
	GameHelpIncreaseStrength             string
	GameHelpDecreaseStrength             string
	GameHelpIgnition                     string
	GameHelpInvertAngle                  string
	GameHelpScrollOMat                   string
	GameHelpIncreaseStrengthByTen        string
	GameHelpDecreaseStrengthByTen        string
	GameHelpPlayerInfo                   string
	GameHelpScoreTable                   string
	GameHelpQuit                         string
	GameHelpAbortRound                   string
	GameHelpPause                        string
	GameHelpTogglePlayerNames            string
	PlayerInfoTitle                      string
	PlayerInfoMissing                    string
	PlayerInfoTankModel                  string
	PlayerInfoStandardTank               string
	PlayerInfoStatus                     string
	PlayerInfoEnergy                     string
	PlayerInfoMoney                      string
	PlayerInfoEnergyShield               string
	PlayerInfoArsenal                    string
	PlayerInfoTrainingAmmo               string
	RoundTransitionPlayed                string
	RoundTransitionRemainingRounds       string
	XMV12MotorOff                        string
	ShopClassA                           string
	ShopClassB                           string
	ShopContinue                         string
	ShopBack                             string
	ShopQuantity                         string
	ShopPrice                            string
	ShopBuy                              string
	ShopBargainBuy                       string
	ShopMoney                            string
	ShopStock                            string
	ItemTrainingAmmo                     string
	ItemGrenade                          string
	ItemLargeGrenade                     string
	ItemAtomBomb                         string
	ItemHBomb                            string
	ItemPlasmaMelter                     string
	ItemWonderPalm                       string
	ItemFireball                         string
	ItemWater                            string
	ItemMoles                            string
	ItemMFSTriple                        string
	ItemSmallCrumblers                   string
	ItemLargeCrumblers                   string
	ItemSurpriseEgg                      string
	ItemMosquitos                        string
	ItemShockwave                        string
	ItemAirStrike                        string
	ItemSplitterBomb                     string
	ItemLaser                            string
	ItemScrollOMat                       string
	ItemEnergyShield                     string
	ItemMFSBooster                       string
	ItemXMV12Tank                        string
	ItemDiesel                           string
	DialogOK                             string
	DialogCancel                         string
}

var currentLanguage = languageGerman

func texts() localizedStrings {
	if currentLanguage == languageEnglish {
		return englishStrings
	}
	return germanStrings
}

func localizedItemName(itemIndex int) string {
	t := texts()
	names := []string{
		t.ItemGrenade,
		t.ItemLargeGrenade,
		t.ItemAtomBomb,
		t.ItemHBomb,
		t.ItemPlasmaMelter,
		t.ItemWonderPalm,
		t.ItemFireball,
		t.ItemWater,
		t.ItemMoles,
		t.ItemMFSTriple,
		t.ItemSmallCrumblers,
		t.ItemLargeCrumblers,
		t.ItemSurpriseEgg,
		t.ItemMosquitos,
		t.ItemShockwave,
		t.ItemAirStrike,
		t.ItemSplitterBomb,
		t.ItemLaser,
		t.ItemScrollOMat,
		t.ItemEnergyShield,
		t.ItemMFSBooster,
		t.ItemXMV12Tank,
		t.ItemDiesel,
	}
	if itemIndex < 0 || itemIndex >= len(names) {
		return ""
	}
	return names[itemIndex]
}

func localizedWeaponSlotName(slot int, fallback string) string {
	t := texts()
	if slot == 0 {
		return t.ItemTrainingAmmo
	}
	name := localizedItemName(slot - 1)
	if name == "" {
		return fallback
	}
	return name
}
