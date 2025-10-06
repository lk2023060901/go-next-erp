package workflow

import (
	"fmt"
	"sort"
	"sync"
)

// ExecutionGraph 执行图（有向无环图 DAG）
// 企业级实现：
// - 完整的 DAG 操作（拓扑排序、环检测、路径查找）
// - 线程安全
// - 高效的图算法实现
// - 支持复杂的依赖关系分析
type ExecutionGraph struct {
	mu sync.RWMutex

	// 节点信息
	nodes map[string]*GraphNode // nodeID -> GraphNode

	// 边信息
	edges     map[string][]*Edge            // source -> edges
	inEdges   map[string][]*Edge            // target -> edges
	adjacency map[string]map[string]bool    // source -> target -> exists

	// 缓存
	topoSortCache [][]string // 拓扑排序缓存
	cacheDirty    bool
}

// GraphNode 图节点
type GraphNode struct {
	ID         string
	Definition *NodeDefinition
	InDegree   int // 入度
	OutDegree  int // 出度
}

// NewExecutionGraph 创建执行图
func NewExecutionGraph() *ExecutionGraph {
	return &ExecutionGraph{
		nodes:     make(map[string]*GraphNode),
		edges:     make(map[string][]*Edge),
		inEdges:   make(map[string][]*Edge),
		adjacency: make(map[string]map[string]bool),
		cacheDirty: true,
	}
}

// AddNode 添加节点
func (g *ExecutionGraph) AddNode(nodeID string, def *NodeDefinition) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.nodes[nodeID]; !exists {
		g.nodes[nodeID] = &GraphNode{
			ID:         nodeID,
			Definition: def,
			InDegree:   0,
			OutDegree:  0,
		}
		g.cacheDirty = true
	}
}

// AddEdge 添加边
func (g *ExecutionGraph) AddEdge(edge *Edge) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// 验证节点存在
	if _, exists := g.nodes[edge.Source]; !exists {
		return fmt.Errorf("source node %s not found", edge.Source)
	}
	if _, exists := g.nodes[edge.Target]; !exists {
		return fmt.Errorf("target node %s not found", edge.Target)
	}

	// 检查是否已存在
	if g.adjacency[edge.Source] == nil {
		g.adjacency[edge.Source] = make(map[string]bool)
	}

	if g.adjacency[edge.Source][edge.Target] {
		return fmt.Errorf("edge already exists: %s -> %s", edge.Source, edge.Target)
	}

	// 添加边
	g.edges[edge.Source] = append(g.edges[edge.Source], edge)
	g.inEdges[edge.Target] = append(g.inEdges[edge.Target], edge)
	g.adjacency[edge.Source][edge.Target] = true

	// 更新节点度数
	g.nodes[edge.Source].OutDegree++
	g.nodes[edge.Target].InDegree++

	g.cacheDirty = true

	return nil
}

// GetNode 获取节点
func (g *ExecutionGraph) GetNode(nodeID string) (*GraphNode, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	node, exists := g.nodes[nodeID]
	return node, exists
}

// GetSuccessors 获取后继节点（出边目标）
func (g *ExecutionGraph) GetSuccessors(nodeID string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	edges := g.edges[nodeID]
	successors := make([]string, 0, len(edges))

	for _, edge := range edges {
		successors = append(successors, edge.Target)
	}

	return successors
}

// GetPredecessors 获取前驱节点（入边来源）
func (g *ExecutionGraph) GetPredecessors(nodeID string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	edges := g.inEdges[nodeID]
	predecessors := make([]string, 0, len(edges))

	for _, edge := range edges {
		predecessors = append(predecessors, edge.Source)
	}

	return predecessors
}

// GetIncomingEdges 获取入边
func (g *ExecutionGraph) GetIncomingEdges(nodeID string) []*Edge {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.inEdges[nodeID]
}

// GetOutgoingEdges 获取出边
func (g *ExecutionGraph) GetOutgoingEdges(nodeID string) []*Edge {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.edges[nodeID]
}

// NodeCount 获取节点数量
func (g *ExecutionGraph) NodeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return len(g.nodes)
}

// EdgeCount 获取边数量
func (g *ExecutionGraph) EdgeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	count := 0
	for _, edges := range g.edges {
		count += len(edges)
	}
	return count
}

// FindStartNodes 查找起始节点（入度为 0）
func (g *ExecutionGraph) FindStartNodes() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var startNodes []string
	for nodeID, node := range g.nodes {
		if node.InDegree == 0 {
			startNodes = append(startNodes, nodeID)
		}
	}

	// 排序以保证确定性
	sort.Strings(startNodes)

	return startNodes
}

// FindEndNodes 查找终止节点（出度为 0）
func (g *ExecutionGraph) FindEndNodes() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var endNodes []string
	for nodeID, node := range g.nodes {
		if node.OutDegree == 0 {
			endNodes = append(endNodes, nodeID)
		}
	}

	// 排序以保证确定性
	sort.Strings(endNodes)

	return endNodes
}

// TopologicalSort 拓扑排序（Kahn 算法）
// 返回：按层级分组的节点 ID 列表
// 同一层级的节点可以并行执行
func (g *ExecutionGraph) TopologicalSort() ([][]string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// 使用缓存
	if !g.cacheDirty && g.topoSortCache != nil {
		return g.topoSortCache, nil
	}

	// 检测环
	if g.hasCycleUnsafe() {
		return nil, ErrCyclicDependency
	}

	// 初始化入度表（复制，避免修改原始数据）
	inDegree := make(map[string]int)
	for nodeID, node := range g.nodes {
		inDegree[nodeID] = node.InDegree
	}

	// 初始化队列（入度为 0 的节点）
	var queue []string
	for nodeID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeID)
		}
	}

	// 排序队列以保证确定性
	sort.Strings(queue)

	// 按层级存储结果
	var layers [][]string
	visited := 0

	for len(queue) > 0 {
		// 当前层级
		currentLayer := queue
		queue = nil

		// 添加到结果
		layers = append(layers, currentLayer)
		visited += len(currentLayer)

		// 下一层级的节点
		nextLayer := make(map[string]bool)

		// 处理当前层级的每个节点
		for _, nodeID := range currentLayer {
			// 减少后继节点的入度
			for _, edge := range g.edges[nodeID] {
				targetID := edge.Target
				inDegree[targetID]--

				// 入度变为 0，加入下一层级
				if inDegree[targetID] == 0 {
					nextLayer[targetID] = true
				}
			}
		}

		// 转换为数组并排序
		if len(nextLayer) > 0 {
			queue = make([]string, 0, len(nextLayer))
			for nodeID := range nextLayer {
				queue = append(queue, nodeID)
			}
			sort.Strings(queue)
		}
	}

	// 验证是否所有节点都被访问
	if visited != len(g.nodes) {
		return nil, ErrCyclicDependency
	}

	// 缓存结果
	g.topoSortCache = layers
	g.cacheDirty = false

	return layers, nil
}

// HasCycle 检测是否存在环（DFS 算法）
func (g *ExecutionGraph) HasCycle() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.hasCycleUnsafe()
}

// hasCycleUnsafe 检测环（不加锁版本）
func (g *ExecutionGraph) hasCycleUnsafe() bool {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for nodeID := range g.nodes {
		if !visited[nodeID] {
			if g.hasCycleDFS(nodeID, visited, recStack) {
				return true
			}
		}
	}

	return false
}

// hasCycleDFS DFS 检测环
func (g *ExecutionGraph) hasCycleDFS(nodeID string, visited, recStack map[string]bool) bool {
	visited[nodeID] = true
	recStack[nodeID] = true

	// 访问所有后继节点
	for _, edge := range g.edges[nodeID] {
		targetID := edge.Target

		// 未访问过，递归检查
		if !visited[targetID] {
			if g.hasCycleDFS(targetID, visited, recStack) {
				return true
			}
		} else if recStack[targetID] {
			// 在递归栈中，发现环
			return true
		}
	}

	// 回溯
	recStack[nodeID] = false
	return false
}

// Validate 验证图的有效性
func (g *ExecutionGraph) Validate() error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// 1. 检查是否为空
	if len(g.nodes) == 0 {
		return fmt.Errorf("graph is empty")
	}

	// 2. 检查是否存在环
	if g.hasCycleUnsafe() {
		return ErrCyclicDependency
	}

	// 3. 检查是否所有节点都可达（从起始节点）
	startNodes := make([]string, 0)
	for nodeID, node := range g.nodes {
		if node.InDegree == 0 {
			startNodes = append(startNodes, nodeID)
		}
	}

	if len(startNodes) == 0 {
		return fmt.Errorf("no start nodes found (all nodes have incoming edges)")
	}

	// 4. 从起始节点 BFS，检查连通性
	reachable := g.bfsReachable(startNodes)
	if len(reachable) != len(g.nodes) {
		unreachable := make([]string, 0)
		for nodeID := range g.nodes {
			if !reachable[nodeID] {
				unreachable = append(unreachable, nodeID)
			}
		}
		return fmt.Errorf("%w: nodes %v are unreachable", ErrDisconnectedGraph, unreachable)
	}

	return nil
}

// bfsReachable BFS 查找所有可达节点
func (g *ExecutionGraph) bfsReachable(startNodes []string) map[string]bool {
	reachable := make(map[string]bool)
	queue := make([]string, len(startNodes))
	copy(queue, startNodes)

	for _, nodeID := range startNodes {
		reachable[nodeID] = true
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, edge := range g.edges[current] {
			if !reachable[edge.Target] {
				reachable[edge.Target] = true
				queue = append(queue, edge.Target)
			}
		}
	}

	return reachable
}

// GetAllPaths 获取从起始节点到终止节点的所有路径
// 用于复杂的依赖分析和调试
func (g *ExecutionGraph) GetAllPaths(start, end string) [][]string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var paths [][]string
	visited := make(map[string]bool)
	currentPath := []string{start}

	g.dfsAllPaths(start, end, visited, currentPath, &paths)

	return paths
}

// dfsAllPaths DFS 查找所有路径
func (g *ExecutionGraph) dfsAllPaths(current, target string, visited map[string]bool, currentPath []string, paths *[][]string) {
	if current == target {
		// 找到一条路径，复制并保存
		path := make([]string, len(currentPath))
		copy(path, currentPath)
		*paths = append(*paths, path)
		return
	}

	visited[current] = true

	for _, edge := range g.edges[current] {
		next := edge.Target
		if !visited[next] {
			g.dfsAllPaths(next, target, visited, append(currentPath, next), paths)
		}
	}

	visited[current] = false
}

// GetLongestPath 获取最长路径长度（用于估算执行时间）
func (g *ExecutionGraph) GetLongestPath() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if len(g.nodes) == 0 {
		return 0
	}

	// 使用动态规划计算最长路径
	distance := make(map[string]int)

	// 拓扑排序
	sorted, err := g.topologicalSortUnsafe()
	if err != nil {
		return 0
	}

	// 展平层级
	var flatSorted []string
	for _, layer := range sorted {
		flatSorted = append(flatSorted, layer...)
	}

	// DP 计算最长路径
	for _, nodeID := range flatSorted {
		maxDist := 0
		for _, edge := range g.inEdges[nodeID] {
			if dist, ok := distance[edge.Source]; ok {
				if dist+1 > maxDist {
					maxDist = dist + 1
				}
			}
		}
		distance[nodeID] = maxDist
	}

	// 找到最大距离
	maxPath := 0
	for _, dist := range distance {
		if dist > maxPath {
			maxPath = dist
		}
	}

	return maxPath + 1 // +1 因为距离是边数，节点数 = 边数 + 1
}

// topologicalSortUnsafe 拓扑排序（不加锁版本）
func (g *ExecutionGraph) topologicalSortUnsafe() ([][]string, error) {
	if !g.cacheDirty && g.topoSortCache != nil {
		return g.topoSortCache, nil
	}

	// 与 TopologicalSort 相同的逻辑，但不加锁
	if g.hasCycleUnsafe() {
		return nil, ErrCyclicDependency
	}

	inDegree := make(map[string]int)
	for nodeID, node := range g.nodes {
		inDegree[nodeID] = node.InDegree
	}

	var queue []string
	for nodeID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeID)
		}
	}

	sort.Strings(queue)

	var layers [][]string
	visited := 0

	for len(queue) > 0 {
		currentLayer := queue
		queue = nil

		layers = append(layers, currentLayer)
		visited += len(currentLayer)

		nextLayer := make(map[string]bool)

		for _, nodeID := range currentLayer {
			for _, edge := range g.edges[nodeID] {
				targetID := edge.Target
				inDegree[targetID]--

				if inDegree[targetID] == 0 {
					nextLayer[targetID] = true
				}
			}
		}

		if len(nextLayer) > 0 {
			queue = make([]string, 0, len(nextLayer))
			for nodeID := range nextLayer {
				queue = append(queue, nodeID)
			}
			sort.Strings(queue)
		}
	}

	if visited != len(g.nodes) {
		return nil, ErrCyclicDependency
	}

	return layers, nil
}

// Clone 克隆图（用于子工作流）
func (g *ExecutionGraph) Clone() *ExecutionGraph {
	g.mu.RLock()
	defer g.mu.RUnlock()

	cloned := NewExecutionGraph()

	// 复制节点
	for nodeID, node := range g.nodes {
		cloned.nodes[nodeID] = &GraphNode{
			ID:         node.ID,
			Definition: node.Definition,
			InDegree:   node.InDegree,
			OutDegree:  node.OutDegree,
		}
	}

	// 复制边
	for source, edges := range g.edges {
		cloned.edges[source] = make([]*Edge, len(edges))
		copy(cloned.edges[source], edges)
	}

	for target, edges := range g.inEdges {
		cloned.inEdges[target] = make([]*Edge, len(edges))
		copy(cloned.inEdges[target], edges)
	}

	// 复制邻接表
	for source, targets := range g.adjacency {
		cloned.adjacency[source] = make(map[string]bool)
		for target := range targets {
			cloned.adjacency[source][target] = true
		}
	}

	cloned.cacheDirty = true

	return cloned
}

// String 返回图的字符串表示（用于调试）
func (g *ExecutionGraph) String() string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return fmt.Sprintf("ExecutionGraph{nodes=%d, edges=%d}", len(g.nodes), g.edgeCountUnsafe())
}

// edgeCountUnsafe 获取边数量（不加锁）
func (g *ExecutionGraph) edgeCountUnsafe() int {
	count := 0
	for _, edges := range g.edges {
		count += len(edges)
	}
	return count
}
