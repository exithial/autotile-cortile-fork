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
	ColumnProps    []float64 // Proportions for each column
}

func CreateAutotileLayout(loc store.Location) *AutotileLayout {
	layout := &AutotileLayout{
		Name:           "autotile",
		Manager:        store.CreateManager(loc),
		Columns:        common.Config.AutotileColumnsDefault,
		ColumnsDefault: common.Config.AutotileColumnsDefault,
		ColumnProps:    make([]float64, common.Config.AutotileColumnsMax),
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
	// Initialize column proportions
	l.ColumnProps = make([]float64, common.Config.AutotileColumnsMax)
	for i := range l.ColumnProps {
		l.ColumnProps[i] = 1.0 / float64(common.Config.AutotileColumnsMax)
	}
}

func (l *AutotileLayout) Apply() {
	clients := l.Clients(store.Stacked)

	dx, dy, dw, dh := store.DesktopGeometry(l.Location.Screen).Pieces()
	gap := common.Config.WindowGapSize

	csize := len(clients)

	cols := l.calculateColumns(csize)

	isUltrawide := dw > common.Config.UltrawideThreshold
	if !isUltrawide && cols > 2 {
		// NOTE: En resoluciones estándar (< ultrawide_threshold), limitar a 2 columnas
		// evita ventanas demasiado estrechas para ser usables. En ultrawide el
		// espacio horizontal es suficiente para hasta AutotileColumnsMax columnas.
		cols = 2
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

	// Pre-calculate column widths using column proportions
	colWidths := make([]int, cols)
	if cols == 1 {
		colWidths[0] = dw
	} else {
		// Calculate total of proportions for visible columns
		totalProp := 0.0
		for i := 0; i < cols; i++ {
			if i < len(l.ColumnProps) {
				totalProp += l.ColumnProps[i]
			} else {
				totalProp += 1.0 / float64(cols)
			}
		}

		// Normalize and calculate widths
		for i := 0; i < cols; i++ {
			prop := 1.0 / float64(cols)
			if i < len(l.ColumnProps) {
				prop = l.ColumnProps[i] / totalProp
			}

			// Ensure minimum proportion
			minProp := common.Config.ProportionMin
			if prop < minProp {
				prop = minProp
			}

			colWidths[i] = int(math.Round(float64(dw) * prop))
		}

		// Adjust for rounding errors
		widthSum := 0
		for i := 0; i < cols; i++ {
			widthSum += colWidths[i]
		}

		if widthSum != dw {
			colWidths[cols-1] += dw - widthSum
		}
	}

	rowsPerCol := make([]int, cols)
	
	// Identificar columna master
	masterCol := l.determineMasterColumn(cols)
	
	if csize <= cols {
		// Menos ventanas que columnas: cada ventana en su columna
		// Intentar poner master en su posición preferida si hay espacio
		for i := 0; i < csize; i++ {
			rowsPerCol[i] = 1
		}
		// Si masterCol está fuera del rango (ej: masterCol=1 pero solo 1 ventana),
		// se quedará en 0, lo cual está bien
	} else {
		// Más ventanas que columnas
		// Primero, 1 ventana por columna
		for i := 0; i < cols; i++ {
			rowsPerCol[i] = 1
		}
		remaining := csize - cols
		
		// Distribuir ventanas restantes, priorizando slaves
		// Queremos que slaves tengan al menos 2 antes de que master tenga 2
		for remaining > 0 {
			// Contar cuántas slaves tienen menos de 2 ventanas
			slavesWithLessThan2 := 0
			for i := 0; i < cols; i++ {
				if i != masterCol && rowsPerCol[i] < 2 {
					slavesWithLessThan2++
				}
			}
			
			if slavesWithLessThan2 > 0 {
				// Hay slaves con menos de 2 ventanas, darles prioridad
				// Encontrar slave con menos ventanas
				minSlaveCol := -1
				minRows := int(^uint(0) >> 1)
				for i := 0; i < cols; i++ {
					if i == masterCol {
						continue
					}
					if rowsPerCol[i] < minRows {
						minRows = rowsPerCol[i]
						minSlaveCol = i
					}
				}
				rowsPerCol[minSlaveCol]++
			} else {
				// Todas las slaves tienen al menos 2, ahora puede dividirse master
				rowsPerCol[masterCol]++
			}
			remaining--
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
			x += gap / 2
			width -= gap / 2
		}
		if col < cols-1 {
			width -= gap / 2
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
				y += gap / 2
				height -= gap / 2
			}
			if row < rows-1 {
				height -= gap / 2
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
	// En autotile, el master depende del número de columnas
	clients := l.Clients(store.Stacked)
	if len(clients) == 0 {
		return
	}
	
	// Encontrar índice del cliente actual
	currentIndex := -1
	for i, client := range clients {
		if client.Window.Id == c.Window.Id {
			currentIndex = i
			break
		}
	}
	if currentIndex == -1 {
		return
	}
	
	// Determinar índice del master según número de columnas
	masterIndex := l.determineMasterIndex()
	if masterIndex >= len(clients) {
		masterIndex = len(clients) - 1
	}
	
	// Intercambiar cliente actual con el master
	if currentIndex != masterIndex {
		clients[currentIndex], clients[masterIndex] = clients[masterIndex], clients[currentIndex]
		// Actualizar listas de masters y slaves
		l.updateMasterSlaveLists(clients)
	}
}
func (l *AutotileLayout) ResetColumns() {
	l.Columns = l.ColumnsDefault
	log.Info("Reset columns to default: ", l.Columns)
}

func (l *AutotileLayout) IncreaseProportion() {
	l.adjustActiveColumnProportion(true)
}

func (l *AutotileLayout) DecreaseProportion() {
	l.adjustActiveColumnProportion(false)
}

func (l *AutotileLayout) adjustActiveColumnProportion(increase bool) {
	active := l.ActiveClient()
	if active == nil {
		return
	}

	// Find which column contains the active window
	col := l.findColumnForClient(active)
	if col < 0 || col >= l.Columns {
		return
	}

	step := common.Config.ProportionStep

	// For edge columns, adjust from the adjacent column
	if col == 0 {
		// Leftmost column - adjust from right neighbor
		if increase {
			l.ColumnProps[col] += step
			l.ColumnProps[col+1] -= step
		} else {
			l.ColumnProps[col] -= step
			l.ColumnProps[col+1] += step
		}
	} else if col == l.Columns-1 {
		// Rightmost column - adjust from left neighbor
		if increase {
			l.ColumnProps[col] += step
			l.ColumnProps[col-1] -= step
		} else {
			l.ColumnProps[col] -= step
			l.ColumnProps[col-1] += step
		}
	} else {
		// Middle column - adjust from both neighbors
		if increase {
			l.ColumnProps[col] += step
			// Distribute reduction between left and right neighbors
			reduction := step / 2.0
			l.ColumnProps[col-1] -= reduction
			l.ColumnProps[col+1] -= reduction
		} else {
			l.ColumnProps[col] -= step
			// Distribute increase between left and right neighbors
			increase := step / 2.0
			l.ColumnProps[col-1] += increase
			l.ColumnProps[col+1] += increase
		}
	}

	// Clamp proportions to valid range
	minProp := common.Config.ProportionMin
	maxProp := 1.0 - common.Config.ProportionMin
	for i := 0; i < l.Columns; i++ {
		if l.ColumnProps[i] < minProp {
			l.ColumnProps[i] = minProp
		} else if l.ColumnProps[i] > maxProp {
			l.ColumnProps[i] = maxProp
		}
	}

	// Normalize to ensure total is 1.0
	l.normalizeColumnProportions()
}

func (l *AutotileLayout) findColumnForClient(c *store.Client) int {
	// Get window geometry
	x, _, _, _ := c.OuterGeometry()
	dx, _, dw, _ := store.DesktopGeometry(l.Location.Screen).Pieces()

	// Calculate relative position
	relX := float64(x-dx) / float64(dw)

	// Determine column based on relative position
	// This assumes columns are arranged left to right
	colWidth := 1.0 / float64(l.Columns)
	for col := 0; col < l.Columns; col++ {
		if relX < float64(col+1)*colWidth {
			return col
		}
	}
	return l.Columns - 1
}

func (l *AutotileLayout) normalizeColumnProportions() {
	if l.Columns <= 1 {
		return
	}

	// Ensure sum is approximately 1.0
	total := 0.0
	for i := 0; i < l.Columns; i++ {
		total += l.ColumnProps[i]
	}

	if total <= 0 || math.Abs(total-1.0) > 0.001 {
		// Normalize to sum to 1.0
		for i := 0; i < l.Columns; i++ {
			l.ColumnProps[i] /= total
		}
	}
}

func (l *AutotileLayout) ResetColumnProportions() {
	// Reset all column proportions to equal distribution
	for i := 0; i < common.Config.AutotileColumnsMax; i++ {
		l.ColumnProps[i] = 1.0 / float64(common.Config.AutotileColumnsMax)
	}
	log.Info("Reset column proportions to equal distribution")
}

func (l *AutotileLayout) determineMasterIndex() int {
	clients := l.Clients(store.Stacked)
	if len(clients) == 0 {
		return 0
	}
	
	// Calcular columnas actuales
	cols := l.calculateColumns(len(clients))
	
	// Determinar columna master
	masterCol := l.determineMasterColumn(cols)
	
	// Calcular distribución usando la misma lógica que applyColumns
	rowsPerCol := make([]int, cols)
	csize := len(clients)
	
	if csize <= cols {
		// Menos ventanas que columnas: cada ventana en su columna
		for i := 0; i < csize; i++ {
			rowsPerCol[i] = 1
		}
	} else {
		// Más ventanas que columnas
		// Primero, 1 ventana por columna
		for i := 0; i < cols; i++ {
			rowsPerCol[i] = 1
		}
		remaining := csize - cols
		
		// Distribuir ventanas restantes, priorizando slaves
		// Queremos que slaves tengan al menos 2 antes de que master tenga 2
		for remaining > 0 {
			// Contar cuántas slaves tienen menos de 2 ventanas
			slavesWithLessThan2 := 0
			for i := 0; i < cols; i++ {
				if i != masterCol && rowsPerCol[i] < 2 {
					slavesWithLessThan2++
				}
			}
			
			if slavesWithLessThan2 > 0 {
				// Hay slaves con menos de 2 ventanas, darles prioridad
				// Encontrar slave con menos ventanas
				minSlaveCol := -1
				minRows := int(^uint(0) >> 1)
				for i := 0; i < cols; i++ {
					if i == masterCol {
						continue
					}
					if rowsPerCol[i] < minRows {
						minRows = rowsPerCol[i]
						minSlaveCol = i
					}
				}
				rowsPerCol[minSlaveCol]++
			} else {
				// Todas las slaves tienen al menos 2, ahora puede dividirse master
				rowsPerCol[masterCol]++
			}
			remaining--
		}
	}
	
	// Calcular índice del master en la lista de clientes
	masterIndex := 0
	for i := 0; i < masterCol; i++ {
		masterIndex += rowsPerCol[i]
	}
	// Tomar la primera ventana de la columna master
	return masterIndex
}

func (l *AutotileLayout) determineMasterColumn(cols int) int {
	// Determinar columna del master según número de columnas
	switch cols {
	case 1:
		return 0
	case 2:
		// Con 2 columnas, master es el de la derecha
		return 1
	case 3:
		// Con 3 columnas, master es el del centro
		return 1
	case 4:
		// Con 4 columnas, master es el del centro-derecha
		return 2
	default:
		// Para más de 4 columnas, usar el del centro
		return cols / 2
	}
}

func (l *AutotileLayout) updateMasterSlaveLists(clients []*store.Client) {
	// Limpiar listas actuales
	l.Masters.Stacked = make([]*store.Client, 0)
	l.Slaves.Stacked = make([]*store.Client, 0)
	
	// Reconstruir listas manteniendo el orden
	for i, client := range clients {
		if i < l.Masters.Maximum {
			l.Masters.Stacked = append(l.Masters.Stacked, client)
		} else {
			l.Slaves.Stacked = append(l.Slaves.Stacked, client)
		}
	}
}
