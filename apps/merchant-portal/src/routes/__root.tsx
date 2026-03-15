import { HeadContent, Scripts, createRootRoute } from '@tanstack/react-router'
import appCss from '../styles.css?url'

// Inline script to prevent FOUC (flash of unstyled content) on page load.
// Runs before React hydration to apply the correct theme class immediately.
const THEME_INIT_SCRIPT = `(function(){try{var stored=localStorage.getItem('theme');var prefersDark=window.matchMedia('(prefers-color-scheme:dark)').matches;var resolved=stored==='dark'?'dark':stored==='light'?'light':prefersDark?'dark':'light';document.documentElement.classList.add(resolved);document.documentElement.style.colorScheme=resolved;}catch(e){}})();`

export const Route = createRootRoute({
  head: () => ({
    meta: [
      { charSet: 'utf-8' },
      { name: 'viewport', content: 'width=device-width, initial-scale=1' },
      { title: 'Open Pay | Merchant Portal' },
      { name: 'description', content: 'Crypto payment processing dashboard for merchants' },
    ],
    links: [
      { rel: 'stylesheet', href: appCss },
    ],
  }),
  shellComponent: RootDocument,
})

function RootDocument({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <script dangerouslySetInnerHTML={{ __html: THEME_INIT_SCRIPT }} />
        <HeadContent />
      </head>
      <body className="min-h-screen bg-background font-sans antialiased" suppressHydrationWarning>
        {children}
        <Scripts />
      </body>
    </html>
  )
}
