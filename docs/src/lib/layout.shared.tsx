import type { BaseLayoutProps } from 'fumadocs-ui/layouts/shared';

export function baseOptions(): BaseLayoutProps {
  return {
    nav: {
      title: (
        <span className="font-bold flex items-center gap-2">
          <span className="bg-gradient-to-r from-indigo-500 to-purple-500 text-white p-1 rounded-lg">
            OR
          </span>
          Octo Router
        </span>
      ),
    },
  };
}
