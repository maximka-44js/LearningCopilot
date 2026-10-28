import { defineConfig } from 'vitepress'
import { generateSidebar } from 'vitepress-sidebar'

export default defineConfig({
  title: "lc_docs",
  description: "documentation",
  base: '/docs/',
  cleanUrls: true,
  vite: {
    server: {
      allowedHosts: ['frontend.lvh.me', 'localhost', '127.0.0.1'],
    },
  },
  themeConfig: {
    // Верхнее навигационное меню
    nav: [
      // { text: 'Платформа', link: '/' },
      { text: 'Документация', link: '/' },
    ],

    // Автоматический сайдбар на основе файловой структуры
    sidebar: generateSidebar({
      documentRootPath: '/',          // Сканируем от корня проекта
      excludeFolders: [
        '.vitepress', 
        'node_modules', 
        'public'
      ],                              // Игнорируем служебные папки
      collapseDepth: 2,               // Уровень вложенности для авто-сворачивания
      capitalizeFirst: true,          // Делать первую букву названий заглавной
      useTitleFromFileHeading: true,  // Брать имя страницы из первого заголовка # в файле
      useFolderTitleFromIndexFile: true // Если в папке есть index.md, имя папки в меню возьмется из него
    }),

    // Иконка гитхаба в правом верхнем углу
    socialLinks: [
      { icon: 'github', link: 'https://github.com' }
    ],

    // Встроенный быстрый поиск по всей доке
    search: {
      provider: 'local'
    }
  }
})
