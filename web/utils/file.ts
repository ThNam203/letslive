export const IsValidFileSizeInMB = (file: File, maxSizeInMB: number) => {
    const fileSizeInMB = file.size / (1024 * 1024); // Convert bytes to MB
    return fileSizeInMB <= maxSizeInMB;
};

// a data URL needs no revoking, unlike URL.createObjectURL
export const readFileAsDataUrl = (file: Blob): Promise<string> =>
    new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = () =>
            typeof reader.result === "string"
                ? resolve(reader.result)
                : reject(new Error("unexpected FileReader result"));
        reader.onerror = () => reject(reader.error);
        reader.readAsDataURL(file);
    });
