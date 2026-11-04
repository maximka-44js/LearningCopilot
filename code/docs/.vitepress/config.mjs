import { defineConfig } from 'vitepress'
import { generateSidebar } from 'vitepress-sidebar'

export default defineConfig({
  title: "Learning Copilot",
  description: "Помощь и техническая документация Learning Copilot",
  base: '/docs/',
  cleanUrls: true,
  vite: {
    server: {
      allowedHosts: ['frontend.lvh.me', 'localhost', '127.0.0.1'],
    },
  },
  themeConfig: {
    nav: [
      { text: 'Помощь', link: '/help' },
      { text: 'Техническая документация', link: '/technical/' },
    ],
    sidebar: generateSidebar({
      documentRootPath: '/',
      excludeFolders: ['.vitepress', 'node_modules', 'public'],
      collapseDepth: 2,
      capitalizeFirst: true,
      useTitleFromFileHeading: true,
      useFolderTitleFromIndexFile: true,
    }),
    socialLinks: [
      { icon: 'github', link: 'https://github.com' },
    ],
    search: {
      provider: 'local',
    },
  },
})
