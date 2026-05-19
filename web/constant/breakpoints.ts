/**
 * Tailwind CSS default min-width breakpoints (px).
 * If you customize `--breakpoint-*` in @theme, update these values to match.
 * @see https://tailwindcss.com/docs/responsive-design
 */
export const BREAKPOINT_SM_PX = 640;
export const BREAKPOINT_MD_PX = 768;
export const BREAKPOINT_LG_PX = 1024;
export const BREAKPOINT_XL_PX = 1280;
export const BREAKPOINT_2XL_PX = 1536;

/** matchMedia query aligned with Tailwind `max-*` (viewport below a min-width breakpoint). */
export function maxWidthBelow(breakpointMinWidthPx: number): string {
    return `(max-width: ${breakpointMinWidthPx - 1}px)`;
}

/** Same viewport as Tailwind `max-md:` */
export const MQ_MAX_MD = maxWidthBelow(BREAKPOINT_MD_PX);

/** Same viewport as Tailwind `max-lg:` */
export const MQ_MAX_LG = maxWidthBelow(BREAKPOINT_LG_PX);
