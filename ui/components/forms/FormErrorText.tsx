export default function FormErrorText({ textError }: { textError: string | undefined }) {
    if (textError) {
        return (
            <p className="text-destructive text-xs font-semibold">
                {textError}
            </p>
        )
    }
}