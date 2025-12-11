package model

// ChartData 图表历史数据
type ChartData struct {
	SpeedUpHistory   []int64 // 上传速度历史 (bytes/s)
	SpeedDownHistory []int64 // 下载速度历史 (bytes/s)
	MemoryHistory    []int64 // 内存使用历史 (bytes)
	ConnCountHistory []int   // 连接数历史
	MaxPoints        int     // 最大数据点数量
}

// NewChartData 创建新的图表数据
func NewChartData(maxPoints int) *ChartData {
	if maxPoints <= 0 {
		maxPoints = 60 // 默认60个数据点
	}
	return &ChartData{
		SpeedUpHistory:   make([]int64, 0, maxPoints),
		SpeedDownHistory: make([]int64, 0, maxPoints),
		MemoryHistory:    make([]int64, 0, maxPoints),
		ConnCountHistory: make([]int, 0, maxPoints),
		MaxPoints:        maxPoints,
	}
}

// AddSpeedData 添加速度数据点
func (c *ChartData) AddSpeedData(upload, download int64) {
	c.SpeedUpHistory = appendWithLimit(c.SpeedUpHistory, upload, c.MaxPoints)
	c.SpeedDownHistory = appendWithLimit(c.SpeedDownHistory, download, c.MaxPoints)
}

// AddMemoryData 添加内存数据点
func (c *ChartData) AddMemoryData(memory int64) {
	c.MemoryHistory = appendWithLimit(c.MemoryHistory, memory, c.MaxPoints)
}

// AddConnCountData 添加连接数数据点
func (c *ChartData) AddConnCountData(count int) {
	c.ConnCountHistory = appendIntWithLimit(c.ConnCountHistory, count, c.MaxPoints)
}

// appendWithLimit 添加数据并限制长度
func appendWithLimit(slice []int64, value int64, maxLen int) []int64 {
	slice = append(slice, value)
	if len(slice) > maxLen {
		slice = slice[len(slice)-maxLen:]
	}
	return slice
}

// appendIntWithLimit 添加int数据并限制长度
func appendIntWithLimit(slice []int, value int, maxLen int) []int {
	slice = append(slice, value)
	if len(slice) > maxLen {
		slice = slice[len(slice)-maxLen:]
	}
	return slice
}
