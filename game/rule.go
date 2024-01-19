// Copyright 2020 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model of a game-specific rule.

package game

type Rule struct {
	Id             int
	RuleNumber     string
	IsTechnical    bool
	IsRankingPoint bool
	Description    string
}

// All rules from the 2022 game that carry point penalties.
var rules = []*Rule{
	{1, "G211", false, false, "Strategies clearly aimed at forcing the opponent ALLIANCE to violate a rule are not in the spirit of FIRST Robotics Competition and not allowed."},
	{2, "G211", true, false, "Strategies clearly aimed at forcing the opponent ALLIANCE to violate a rule are not in the spirit of FIRST Robotics Competition and not allowed. If REPEATED, TECH FOUL."},
	{3, "G301", true, false, "DRIVE TEAMS may not cause significant delays to the start of their MATCH."},
	{4, "G401", false, false, "During AUTO, DRIVE TEAM members in ALLIANCE AREAS and HUMAN PLAYERS in their SUBSTATION AREAS may not contact anything in front of the STARTING LINES, unless for personal or equipment safety or granted permission by a Head REFEREE or FTA."},
	{5, "G402", false, false, "During AUTO, DRIVE TEAMS may not directly or indirectly interact with ROBOTS or OPERATOR CONSOLES unless for personal safety, OPERATOR CONSOLE safety, or pressing an E-Stop."},
	{6, "G403", true, false, "In AUTO, ROBOTS may not CONTROL more than 1 NOTE at a time, either directly or transitively through other objects."},
	{7, "G404", true, false, "In AUTO, a ROBOT whose BUMPERS are completely outside their WING may	not cause a NOTE to travel into or through their WING such that the NOTE enters the WING while not in contact with that ROBOT."},
	{8, "G405", true, false, "In AUTO, a ROBOT whose BUMPERS have completely crossed the CENTER LINE (i.e. to the opposite side of the CENTER LINE from its ROBOT STARTING ZONE) may contact neither an opponent ROBOT nor NOTES staged in the opponent’s WING."},
	{9, "G406", true, false, "ROBOTS may not deliberately use GAME PIECES in an attempt to ease or amplify challenges associated with FIELD elements."},
	{10, "G407", true, false, "ROBOTS may not intentionally eject NOTES from the FIELD (either directly or by bouncing off a FIELD element or other ROBOT) other than through their SPEAKER or AMP."},
	{11, "G408", true, false, "ROBOTS may not cause HIGH NOTES to leave the FIELD (including through an AMP or SPEAKER), score on a MICROPHONE, or enter a TRAP."},
	{12, "G409", false, false, "In TELEOP, a ROBOT may neither A. leave its SOURCE ZONE with CONTROL of more than 1 NOTE nor B. have greater-than-MOMENTARY CONTROL of more than 1 NOTE, either directly or transitively through other objects, while outside their SOURCE ZONE."},
	{13, "G410", true, false, "ROBOTS and HUMAN PLAYERS may not damage GAME PIECES. TECH FOUL if REPEATED."},
	{14, "G412", false, false, "BUMPERS must be in Bumper Zone (see R402) during the match."},
	{15, "G413", false, false, "A ROBOT may not expand beyond either of the following limits: A. its height, as measured when it’s resting normally on a flat floor, may not exceed 4 ft. (~122 cm) or B. it may not extend more than 1 ft. (~30 cm) from its FRAME PERIMETER. Overexpansion due to damage, provided the expansion isn’t leveraged for strategic benefit, is an exception to this rule."},
	{16, "G413", true, false, "A ROBOT may not expand beyond either of the following limits: A. its height, as measured when it’s resting normally on a flat floor, may not exceed 4 ft. (~122 cm) or B. it may not extend more than 1 ft. (~30 cm) from its FRAME PERIMETER. Overexpansion due to damage, provided the expansion isn’t leveraged for strategic benefit, is an exception to this rule. TECH FOUL if the over-expansion impedes or enables a scoring action."},
	{17, "G414", false, false, "A ROBOT with any part of its BUMPERS in their opponent’s WING may not cause NOTES to travel into or through their WING."},
	{18, "G414", true, false, "A ROBOT with any part of its BUMPERS in their opponent’s WING may not cause NOTES to travel into or through their WING. TECH FOUL for subsequent violations in the MATCH."},
	{19, "G415", true, false, "ROBOTS are prohibited from the following interactions with ARENA elements, except chain (see G416) and GAME PIECES (see Section 7.4.2 GAME PIECES). A. grabbing, B. grasping, C. attaching to (including the use of a vacuum or hook fastener to anchor to the FIELD carpet), D. becoming entangled with, and E. suspending from."},
	{20, "G416", true, false, "A ROBOT may not reduce the working length of chain. Incidental actions such as minor twisting due to ROBOT imbalance or ROBOT-to-ROBOT interaction are not considered violations of this rule."},
	{21, "G417", false, false, "A ROBOT may not use a COMPONENT outside its FRAME PERIMETER (except its BUMPERS) to initiate contact with an opponent ROBOT inside the vertical projection of that opponent ROBOT’S FRAME PERIMETER."},
	{22, "G418", true, false, "A ROBOT may not damage or functionally impair an opponent ROBOT in either of the following ways: A. deliberately, as perceived by a REFEREE. B. regardless of intent, by initiating contact, either directly or transitively via a NOTE CONTROLLED by the ROBOT, inside the vertical projection of an opponent ROBOT’S FRAME PERIMETER. Damage or functional impairment because of contact with a tipped-over opponent ROBOT, which is not perceived by a REFEREE to be deliberate, is not a violation of this rule."},
	{23, "G419", true, false, "A ROBOT may not deliberately, as perceived by a REFEREE, attach to, tip, or entangle with an opponent ROBOT."},
	{24, "G420", false, false, "ROBOTS may not PIN an opponent’s ROBOT for more than 5 seconds."},
	{25, "G420", true, false, "ROBOTS may not PIN an opponent’s ROBOT for more than 5 seconds. An additional TECH FOUL for every 5 seconds in which the situation is not corrected."},
	{26, "G421", true, false, "2 or more ROBOTS that appear to a REFEREE to be working together may not isolate or close off any major element of MATCH play."},
	{27, "G422", true, false, "Prior to the last 20 seconds of a MATCH, a ROBOT may not contact (either directly or transitively through a NOTE and regardless of who initiates contact) an opponent ROBOT whose BUMPERS are in contact with their PODIUM."},
	{28, "G423", true, false, "A ROBOT may not contact (either directly or transitively through a NOTE and regardless of who initiates contact) an opponent ROBOT if any part of either ROBOT’S BUMPERS are in the opponent’s SOURCE ZONE or AMP ZONE."},
	{29, "G424", true, true, "A ROBOT may not contact (either directly or transitively through a NOTE and regardless of who initiates contact) an opponent ROBOT if either of the following criteria are met: A. the opponent ROBOT has any part of its BUMPERS in its STAGE ZONE and it is not in contact with the carpet or B. any part of either ROBOT’S BUMPERS are in the opponent’s STAGE ZONE during the last 20 seconds of the MATCH."},
	{30, "G425", false, false, "DRIVE TEAMS must remain in their designated areas."},
	{31, "G426", true, false, "A ROBOT shall be operated only by the DRIVERS and/or HUMAN PLAYERS of that team. A COACH activating their E-Stop or A-Stop is the exception to this rule."},
	{32, "G427", false, false, "DRIVE TEAM members may not extend into the CHUTE."},
	{33, "G428", true, false, "DRIVE TEAM members may not deliberately use GAME PIECES in an attempt to ease or amplify challenges associated with FIELD elements"},
	{34, "G429", true, false, "NOTES may only be introduced to the FIELD through the SOURCE."},
	{35, "G430", false, false, "HIGH NOTES may only be entered on to the FIELD during the last 20 seconds of the MATCH by a HUMAN PLAYER in front of the COACH LINE."},
}
var ruleMap map[int]*Rule

// Returns the rule having the given ID, or nil if no such rule exists.
func GetRuleById(id int) *Rule {
	return GetAllRules()[id]
}

// Returns a slice of all defined rules that carry point penalties.
func GetAllRules() map[int]*Rule {
	if ruleMap == nil {
		ruleMap = make(map[int]*Rule, len(rules))
		for _, rule := range rules {
			ruleMap[rule.Id] = rule
		}
	}
	return ruleMap
}
