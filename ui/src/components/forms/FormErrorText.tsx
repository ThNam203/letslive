export default function FormErrorText({
    textError,
}: {
    textError: string | undefined;
}) {
    if (textError) {
        return (
            <p className="text-xs font-semibold text-destructive">
                {textError}
            </p>
        );
    }
}
