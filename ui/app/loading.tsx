export default function Loading() {
    return (
        <div className="min-h-screen w-full flex flex-col items-center justify-center p-4">
            <div className="max-w-[600px] text-center flex items-center justify-center">
                <div className="w-40 h-40 bg-gray-200 rounded-full mb-6" />
                <div>
                    <h1 className="text-4xl font-bold mb-3">Loading...</h1>
                    <p className="text-muted-foreground text-lg mb-6">
                        Please wait while we load the content
                    </p>
                </div>
            </div>
        </div>
    );
}