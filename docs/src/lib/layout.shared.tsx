import type { BaseLayoutProps } from 'fumadocs-ui/layouts/shared';


export function baseOptions(): BaseLayoutProps {
  return {
    nav: {
      title: (
        <span className="font-bold flex items-center gap-2">
          <img src="/octorouter-logo.svg" alt="Octo Router logo" className="h-6 w-auto" />
          OctoRouter
        </span>
      ),
    },
  };
}
