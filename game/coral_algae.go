package game

import "math"

type Row int

const (
	rowBottom Row = iota
	rowMiddle
	rowTop
	rowTrough
	rowCount
)

type Node int

const (
	nodeA Node = iota
	nodeB
	nodeC
	nodeD
	nodeE
	nodeF
	nodeG
	nodeH
	nodeI
	nodeJ
	nodeK
	nodeL
	nodeCount
)

var autoPoints = map[Row]int{
	rowBottom: 4,
	rowMiddle: 6,
	rowTop:    7,
	rowTrough: 3,
}

var teleopPoints = map[Row]int{
	rowBottom: 3,
	rowMiddle: 4,
	rowTop:    5,
	rowTrough: 2,
}

type Reef struct {
	AutoScoring      [3][12]bool
	Nodes            [3][12]bool
	Algaes           [2][12]bool
	AutoCoralTrough  int
	TotalCoralTrough int
}

type CoralAlgae struct {
	Reef
	AutoProcessorAlgae   int
	TeleopProcessorAlgae int
	AutoNetAlgae         int
	TeleopNetAlgae       int
}

func (coralAlgae *CoralAlgae) AutoAlgaePoints() int {
	return coralAlgae.AutoProcessorAlgae*6 + coralAlgae.AutoNetAlgae*4
}

func (coralAlgae *CoralAlgae) TeleopAlgaePoints() int {
	return coralAlgae.TeleopProcessorAlgae*6 + coralAlgae.TeleopNetAlgae*4
}

func (coralAlgae *CoralAlgae) AutoCoralPoints() int {
	points := 0
	for row := rowBottom; row < rowTrough; row++ {
		for node := nodeA; node < nodeCount; node++ {
			autoCoral, _ := coralAlgae.numScoredAutoTeleopCoral(row, node)
			points += autoCoral * autoPoints[row]
		}
	}
	autoCoral, _ := coralAlgae.numScoredAutoTeleopCoral(rowTrough, nodeA)
	points += autoCoral * autoPoints[rowTrough]

	return points
}

func (coralAlgae *CoralAlgae) TeleopCoralPoints() int {
	points := 0
	for row := rowBottom; row < rowTrough; row++ {
		for node := nodeA; node < nodeCount; node++ {
			_, teleopCoral := coralAlgae.numScoredAutoTeleopCoral(row, node)
			points += teleopCoral * teleopPoints[row]
		}
	}
	_, teleopCoral := coralAlgae.numScoredAutoTeleopCoral(rowTrough, nodeA)
	points += teleopCoral * teleopPoints[rowTrough]

	return points
}

func (coralAlgae *CoralAlgae) NumScoredCoralEachRow() [4]int {
	var coral [4]int
	for row := rowBottom; row < rowTrough; row++ {
		for node := nodeA; node < nodeCount; node++ {
			coral[row] = coralAlgae.numScoredCoral(row, node)
		}
	}
	coral[rowTrough] = coralAlgae.numScoredCoral(rowTrough, nodeA)

	return coral
}

func (coralAlgae *CoralAlgae) IsCoopertitionThresholdAchieved() bool {
	return coralAlgae.AutoProcessorAlgae+coralAlgae.TeleopProcessorAlgae >= 2
}

func (coralAlgae *CoralAlgae) numScoredAutoTeleopCoral(row Row, node Node) (int, int) {
	if row < rowBottom || row > rowTrough || node < nodeA || node > nodeL {
		return 0, 0
	}

	if row == rowTrough {
		return coralAlgae.AutoCoralTrough, int(math.Max(0, float64(coralAlgae.TotalCoralTrough-coralAlgae.AutoCoralTrough)))
	}

	autoScoring := coralAlgae.AutoScoring[row][node]
	nodeState := coralAlgae.Nodes[row][node]

	var autoCoral, teleopCoral int
	autoCoral = 0
	teleopCoral = 0
	if autoScoring && nodeState {
		autoCoral = 1
	} else if !autoScoring && nodeState {
		teleopCoral = 1
	}

	return autoCoral, teleopCoral
}

func (coralAlgae *CoralAlgae) numScoredCoral(row Row, node Node) int {
	autoCoral, teleopCoral := coralAlgae.numScoredAutoTeleopCoral(row, node)
	return autoCoral + teleopCoral
}
