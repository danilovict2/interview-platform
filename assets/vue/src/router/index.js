import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import Cookies from 'js-cookie'
import { useAuthStore } from '@/stores/auth'
const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        {
            path: '/',
            name: 'home',
            component: HomeView,
            meta: {
                requiresAuth: true
            }
        },
        {
            path: '/register',
            name: 'register',
            component: () => import('../views/RegisterView.vue'),
        },
        {
            path: '/login',
            name: 'login',
            component: () => import('../views/LoginView.vue'),
        },
    ],
})

router.beforeEach((to, from, next) => {
    const jwt = Cookies.get('jwt');
    const authStore = useAuthStore();
    authStore.loadAuthUser();

    if (to.meta.requiresAuth && !jwt) {
        next('/login');
    } else if (!to.meta.requiresAuth && jwt) {
        next('/');
    } else {
        next();
    }
})

export default router