import { createRouter, createWebHistory } from "vue-router"
import HomeView from "@/views/HomeView.vue"
import TapeView from "@/views/TapeView.vue"
import SourcesView from "@/views/SourcesView.vue"
import SourceView from "@/views/SourceView.vue"
import NewTapeView from "@/views/NewTapeView.vue"
import Settings from "@/views/Settings.vue"

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "home",
      component: HomeView
    },
    {
      path: "/tapes/new",
      name: "tape-new",
      component: NewTapeView
    },
    {
      path: "/tapes/:tapeId",
      name: "tape",
      component: TapeView
    },
    {
      path: "/sources",
      name: "sources",
      component: SourcesView
    },
    {
      path: "/sources/:sourceId",
      name: "source",
      component: SourceView
    },
    {
      path: "/settings",
      name: "settings",
      component: Settings
    },
  ]
})

export default router
