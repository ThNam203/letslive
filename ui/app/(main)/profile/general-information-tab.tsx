import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { User } from "@/types/user";
import { UpdateProfile } from "@/lib/api/user";
import { toast } from "react-toastify";

export default function GeneralInfoTab({ user, updateUser }: { user: User | undefined, updateUser: (user: User) => void }) {
    const [username, setUsername] = useState(user ? user.username : "");
    const [bio, setBio] = useState(user ? user.bio : "");

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!user) return;

        const updatedUserResponse = await UpdateProfile({id: user.id, username, bio });
        if (updatedUserResponse.fetchError) {
            toast.error(updatedUserResponse.fetchError.message);
            return
        }

        setUsername(updatedUserResponse.user!.username);
        setBio(updatedUserResponse.user!.bio);

        updateUser({
            ...user,
            username: updatedUserResponse.user!.username,
            bio: updatedUserResponse.user!.bio,
        });

        toast.success("Profile updated successfully");
    };

    return (
        <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
                <Label htmlFor="name">Name</Label>
                <Input
                    id="name"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    placeholder="Your name"
                />
            </div>
            <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <Input
                    id="email"
                    type="email"
                    value={user ? user.email : ""}
                    readOnly
                    placeholder="Your email"
                    className="bg-gray-100"
                />
            </div>
            <div className="space-y-2">
                <Label htmlFor="bio">Bio</Label>
                <Textarea
                    id="bio"
                    value={bio}
                    onChange={(e) => setBio(e.target.value)}
                    placeholder="Tell us about yourself"
                />
            </div>
            <Button type="submit">Save Changes</Button>
        </form>
    );
}
