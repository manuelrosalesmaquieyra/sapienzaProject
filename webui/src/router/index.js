import {createRouter, createWebHashHistory} from 'vue-router'
import LoginView from '../views/LoginView.vue'
import HomeView from '../views/HomeView.vue'
import ChatView from '../views/ChatView.vue'

const router = createRouter({
	history: createWebHashHistory(import.meta.env.BASE_URL),
	routes: [
		{path: '/', redirect: '/login'},
		{path: '/login', component: LoginView},
		{path: '/home', component: HomeView},
		{ 
			path: '/conversations/:conversation_id', 
			name: 'chat',
			component: ChatView 
		},
		//{path: '/some/:id/link', component: HomeView},
	]
})

export default router
