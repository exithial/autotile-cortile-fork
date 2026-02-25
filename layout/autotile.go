package layout

import (
	"math"

	"github.com/leukipp/cortile/v2/common"
	"github.com/leukipp/cortile/v2/store"

	log "github.com/sirupsen/logrus"
)

type AutotileLayout struct {
	Name string
	*store.Manager
	Columns        int
	ColumnsDefault int
}

func CreateAutotileLayout(loc store.Location) *AutotileLayout {
	layout := &AutotileLayout{
		Name:           "autotile",
		Manager:        store.CreateManager(loc),
		Columns:        common.Config.AutotileColumnsDefault,
		ColumnsDefault: common.Config.AutotileColumnsDefault,
	}
	layout.Reset()
	return layout
}

func (l *AutotileLayout) Reset() {
	mg := store.CreateManager(*l.Location)

	for l.Masters.Maximum < mg.Masters.Maximum {
		l.IncreaseMaster()
	}
	for l.Masters.Maximum > mg.Masters.Maximum {
		l.DecreaseMaster()
	}

	for l.Slaves.Maximum < mg.Slaves.Maximum {
		l.IncreaseSlave()
	}
	for l.Slaves.Maximum > mg.Slaves.Maximum {
		l.DecreaseSlave()
	}

	l.Manager.Proportions = mg.Proportions
	l.Columns = l.ColumnsDefault
}

func (l *AutotileLayout) Apply() {
	clients := l.Clients(store.Stacked)

	dx, dy, dw, dh := store.DesktopGeometry(l.Location.Screen).Pieces()
	gap := common.Config.WindowGapSize

	csize := len(clients)

	isUltrawide := dw > common.Config.UltrawideThreshold
	cols := l.Columns
	if isUltrawide {
		cols = l.calculateColumns(csize)
	}

	if cols < 1 {
		cols = 1
	}
	if cols > common.Config.AutotileColumnsMax {
		cols = common.Config.AutotileColumnsMax
	}

	log.Info("Tile ", csize, " windows with ", l.Name, " layout (", cols, " columns) [workspace-", l.Location.Desktop, "-", l.Location.Screen, "]")

	l.applyColumns(dx, dy, dw, dh, gap, cols, csize)
}

func (l *AutotileLayout) calculateColumns(clientCount int) int {
	if clientCount <= 1 {
		return 1
	}
	if clientCount <= l.Columns {
		return clientCount
	}
	return l.Columns
}

func (l *AutotileLayout) applyColumns(dx, dy, dw, dh, gap, cols, csize int) {
	if csize == 0 {
		return
	}

	// Pre-calculate column widths
	colWidths := make([]int, cols)
	for i := 0; i < cols; i++ {
		colWidths[i] = dw / cols
		if i < dw % cols {
			colWidths[i]++
		}
	}

	rowsPerCol := make([]int, cols)
	for i := 0; i < cols; i++ {
		rowsPerCol[i] = csize / cols
		if i < csize%cols {
			rowsPerCol[i]++
		}
	}

	currentClient := 0
	for col := 0; col < cols; col++ {
		rows := rowsPerCol[col]
		if rows == 0 {
			continue
		}

		// Calculate column position and width
		x := dx
		for i := 0; i < col; i++ {
			x += colWidths[i]
		}

		width := colWidths[col]

		// Apply gap/2 between columns
		if col > 0 {
			x += gap/2
			width -= gap/2
		}
		if col < cols-1 {
			width -= gap/2
		}

			// Calculate available height for this column
			heightTotal := dh
			rowHeight := heightTotal / rows
			rowRemainder := heightTotal % rows

			// Pre-calculate row heights
			rowHeights := make([]int, rows)
			for i := 0; i < rows; i++ {
				rowHeights[i] = rowHeight
				if i < rowRemainder {
					rowHeights[i]++
				}
			}

			for row := 0; row < rows; row++ {
				if currentClient >= csize {
					break
				}

				c := l.Clients(store.Stacked)[currentClient]
				if c == nil {
					currentClient++
					continue
				}

				// Calculate row position
				y := dy
				for i := 0; i < row; i++ {
					y += rowHeights[i]
				}

				// Row height
				height := rowHeights[row]

				// Apply gap/2 between rows
				if row > 0 {
					y += gap/2
					height -= gap/2
				}
				if row < rows-1 {
					height -= gap/2
				}

			// Apply outer gaps
			xPos := x
			yPos := y
			w := width
			h := height

			if col == 0 {
				xPos += gap
				w -= gap
			}
			if col == cols-1 {
				w -= gap
			}
			if row == 0 {
				yPos += gap
				h -= gap
			}
			if row == rows-1 {
				h -= gap
			}

			// Limit minimum dimensions
			minw := int(math.Round(float64(dw-2*gap) * common.Config.ProportionMin))
			minh := int(math.Round(float64(dh-2*gap) * common.Config.ProportionMin))
			c.Limit(minw, minh)

			c.MoveWindow(xPos, yPos, w, h)

			currentClient++
		}
	}
}

	func (l *AutotileLayout) UpdateProportions(c *store.Client, d *store.Directions) {
	_, _, dw, dh := store.DesktopGeometry(l.Location.Screen).Pieces()
	_, _, cw, ch := c.OuterGeometry()

	gap := common.Config.WindowGapSize

	px := float64(cw+gap) / float64(dw)
	py := float64(ch+gap) / float64(dh)

	l.Manager.SetProportions(l.Proportions.MasterSlave[2], px, 0, 1)

	if d.Left {
		l.Manager.SetProportions(l.Proportions.MasterSlave[2], px, 0, 1)
	} else if d.Right {
		l.Manager.SetProportions(l.Proportions.MasterSlave[2], px, 1, 0)
	}

	if d.Top {
		l.Manager.SetProportions(l.Proportions.MasterSlave[2], py, 0, 1)
	} else if d.Bottom {
		l.Manager.SetProportions(l.Proportions.MasterSlave[2], py, 1, 0)
	}
}

	func (l *AutotileLayout) IncreaseColumn() {
	if l.Columns < common.Config.AutotileColumnsMax {
		l.Columns++
		log.Info("Increase columns to ", l.Columns)
	}
}

	func (l *AutotileLayout) DecreaseColumn() {
	if l.Columns > 1 {
		l.Columns--
		log.Info("Decrease columns to ", l.Columns)
	}
}

	func (l *AutotileLayout) GetManager() *store.Manager {
	return l.Manager
}

	func (l *AutotileLayout) GetName() string {
	return l.Name
}

	func (l *AutotileLayout) IncreaseSlave() {
	l.Manager.IncreaseSlave()
}

	func (l *AutotileLayout) DecreaseSlave() {
	l.Manager.DecreaseSlave()
}

	func (l *AutotileLayout) IncreaseMaster() {
	l.Manager.IncreaseMaster()
}

	func (l *AutotileLayout) DecreaseMaster() {
	l.Manager.DecreaseMaster()
}

	func (l *AutotileLayout) AddClient(c *store.Client) {
	l.Manager.AddClient(c)
}

	func (l *AutotileLayout) RemoveClient(c *store.Client) {
	l.Manager.RemoveClient(c)
}

	func (l *AutotileLayout) ActiveClient() *store.Client {
	return l.Manager.ActiveClient()
}

	func (l *AutotileLayout) NextClient() *store.Client {
	return l.Manager.NextClient()
}

	func (l *AutotileLayout) PreviousClient() *store.Client {
	return l.Manager.PreviousClient()
}

	func (l *AutotileLayout) MakeMaster(c *store.Client) {
	l.Manager.MakeMaster(c)
}
func (l *AutotileLayout) ResetColumns() {
	l.Columns = l.ColumnsDefault
	log.Info("Reset columns to default: ", l.Columns)
}
