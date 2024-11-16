import {type DefaultTheme, defineConfig} from 'vitepress'
import { withMermaid } from "vitepress-plugin-mermaid";

export default withMermaid(defineConfig({
    title: "Flash (Quix Labs)",
    lang: 'en-US',
    description: "A lightweight Go library for tracking and managing real-time PostgreSQL changes seamlessly and efficiently.",

    lastUpdated: false,
    cleanUrls: true,

    srcExclude: [
        'README.md'
    ],

    head: [
        ['link', {rel: 'icon', type: 'image/svg+xml', href: '/logo.svg'}],
        ['link', {rel: 'icon', type: 'image/png', sizes: '32x32', href: '/favicon-32x32.png'}],
        ['link', {rel: 'icon', type: 'image/png', sizes: '16x16', href: '/favicon-16x16.png'}],
        ['link', {rel: 'apple-touch-icon', sizes: '180x180', href: '/apple-touch-icon.png'}],
        ['meta', {name: 'theme-color', content: '#5f67ee'}],
        ['meta', {property: 'og:type', content: 'website'}],
        ['meta', {property: 'og:locale', content: 'en'}],
        ['meta', {property: 'og:title', content: 'Flash | Keep track of your database changes'}],
        ['meta', {property: 'twitter:title', content: 'Flash | Keep track of your database changes'}],
        ['meta', {property: 'og:site_name', content: 'Flash'}],
        ['meta', {property: 'twitter:card', content: 'summary_large_image'}],
        ['meta', {property: 'twitter:image:src', content: 'https://flash.quix-labs.com/flash-og.png'}],
        ['meta', {property: 'og:image', content: 'https://flash.quix-labs.com/flash-og.png'}],
        ['meta', {property: 'og:image:type', content: 'image/png'}],
        ['meta', {property: 'og:image:width', content: '1280'}],
        ['meta', {property: 'og:image:height', content: '640'}],
        ['meta', {property: 'og:url', content: 'https://flash.quix-labs.com'}],
    ],

    sitemap: {
        hostname: 'https://flash.quix-labs.com'
    },

    themeConfig: {
        outline: [2, 3],
        logo: '/logo.svg',
        siteTitle: "Flash",
        nav: [
            {text: 'Guide', link: '/guide/what-is-flash', activeMatch: '/guide/'},
            {text: 'Team', link: '/team', activeMatch: '/team/'},
        ],

        socialLinks: [
            {icon: 'github', link: 'https://github.com/quix-labs/flash'}
        ],

        sidebar: {
            '/guide/': {base: '/guide/', items: sidebarGuide()},
        },

        editLink: {
            pattern: 'https://github.com/quix-labs/flash/edit/main/docs/:path',
            text: 'Edit this page on GitHub'
        },

        search: {
            provider: 'local',
        },

        footer: {
            message: 'Released under the <a href="https://github.com/quix-labs/flash/blob/main/LICENSE.md">MIT License</a>.',
            copyright: `Copyright © ${new Date().getFullYear()} - <a href="https://www.quix-labs.com">Quix Labs</a>`
        }
    }
}))


function sidebarGuide(): DefaultTheme.SidebarItem[] {
    return [
        {
            text: 'Getting Started',
            collapsed: false,
            items: [
                {text: 'Introduction', link: 'what-is-flash'},
                {text: 'Installation', link: 'installation'},
            ]
        },

        {
            text: 'Usage',
            collapsed: false,
            items: [
                {text: 'Start listening', link: 'start-listening'},
                {text: 'Advanced Features', link: 'advanced-features'},
                {text: 'Drivers Overview', link: 'drivers/'},
            ]
        },

        {
            text: 'Drivers',
            collapsed: false,
            base: '/guide/drivers/',
            items: [
                {text: 'Trigger', link: 'trigger/'},
                {text: 'WAL Logical', link: 'wal_logical/',},
            ]
        },

        {
            text: "Additional Resources",
            collapsed: false,
            items: [
                {text: 'Planned Features', link: 'planned-features'},
                {text: 'Upgrade', link: 'upgrade'},
                {text: 'Contributing Guide', link: 'contributing'},
            ]
        },
    ]
}