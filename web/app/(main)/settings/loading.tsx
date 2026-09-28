import { FormSkeleton } from "@/components/skeletons/form-skeleton";

export default function SettingsLoading() {
    return <FormSkeleton fields={3} multiline={[1]} />;
}
