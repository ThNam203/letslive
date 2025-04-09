export const IsValidFileSizeInMB = (file: File, maxSizeInMB: number) => {
  const fileSizeInMB = file.size / (1024 * 1024); // Convert bytes to MB
  return fileSizeInMB <= maxSizeInMB;
};
