import { createRouter, createWebHistory } from "vue-router"
import HomeView from "@/views/HomeView.vue"
import TapeView from "@/views/TapeView.vue"
import PlaylistView from "@/views/PlaylistView.vue"

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "home",
      component: HomeView
    },
    {
      path: "/tapes/:tapeId",
      component: TapeView
    },
    {
      path: "/playlists/:playlistId",
      name: "playlist",
      component: PlaylistView
    }
  ]
})

export default router
