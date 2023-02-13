export const formatViewsCount = (viewCount: number): string => {
  return Intl.NumberFormat("en-US", { notation: "compact" }).format(viewCount);
};
