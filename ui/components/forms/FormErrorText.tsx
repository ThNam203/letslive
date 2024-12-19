export default function FormErrorText({ textError }: { textError: string | undefined }) {
    if (textError) {
        return (
            <p className="text-red-500 text-xs font-semibold">
                {textError}
            </p>
        )
    }
}