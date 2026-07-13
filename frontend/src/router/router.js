import { createRouter, createWebHashHistory } from 'vue-router'
import DecisionWorkbench from '../components/DecisionWorkbench.vue'

const stockView = () => import('../components/stock.vue')
const settingsView = () => import('../components/settings.vue')
const fundView = () => import('../components/fund.vue')
const marketView = () => import('../components/market.vue')
const agentChat = () => import('../components/agent-chat.vue')
const research = () => import('../components/researchIndex.vue')
const cronTaskManager = () => import('../components/cron-task-manager.vue')
const mcpServerManager = () => import('../components/mcp-server-manager.vue')
const klineAnalysis = () => import('../components/kline-analysis.vue')
const ScreeningHub = () => import('../components/ScreeningHub.vue')
const IntradayRadar = () => import('../components/IntradayRadar.vue')
const HoldingCenter = () => import('../components/HoldingCenter.vue')
const AIConsole = () => import('../components/AIConsole.vue')
const AboutWorkspace = () => import('../components/AboutWorkspace.vue')

const routes = [
	{ path: '/', component: DecisionWorkbench, name: 'dashboard' },
	{ path: '/watchlist', component: stockView, name: 'stock' },
	{ path: '/fund', component: fundView, name: 'fund' },
	{ path: '/settings', component: settingsView, name: 'settings' },
	{ path: '/market', component: marketView, name: 'market' },
	{ path: '/agent', component: agentChat, name: 'agent' },
	{ path: '/research', component: research, name: 'research' },
	{ path: '/cron-tasks', component: cronTaskManager, name: 'cronTasks' },
	{ path: '/mcp-servers', component: mcpServerManager, name: 'mcpServers' },
	{ path: '/kline-analysis', component: klineAnalysis, name: 'klineAnalysis' },
	{ path: '/screening', component: ScreeningHub, name: 'screening' },
	{ path: '/radar', component: IntradayRadar, name: 'radar' },
	{ path: '/holding-center', component: HoldingCenter, name: 'holdingCenter' },
	{ path: '/ai-console', component: AIConsole, name: 'aiConsole' },
	{ path: '/about', component: AboutWorkspace, name: 'about' },
]

const router = createRouter({
	history: createWebHashHistory(),
	routes,
})

export default router
