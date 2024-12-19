"use client";

import { useEffect, useState } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import GeneralInfoTab from "./general-information-tab";
import ChangePasswordTab from "./change-password-tab";
import ApiKeyTab from "./api-key-tab";
import { GetMeProfile } from "@/lib/api/user";
import { toast } from "react-toastify";
import { User } from "@/types/user";

export default function ProfilePage() {
    const [activeTab, setActiveTab] = useState("general");
    const [user, setUser] = useState<User | undefined>(undefined);

    useEffect(() => {
        const fetchProfile = async () => {
            const { user, fetchError } = await GetMeProfile();

            if (fetchError != undefined) {
                toast.error(fetchError.message, {
                    toastId: "profile-fetch-error",
                });
            } else {
                setUser(user);
            }
        };

        fetchProfile();
    }, []);

    return (
        <div className="container mx-auto py-10">
            <Card>
                <CardHeader>
                    <CardTitle>Profile Settings</CardTitle>
                </CardHeader>
                <CardContent>
                    <Tabs value={activeTab} onValueChange={setActiveTab}>
                        <TabsList className="grid w-full grid-cols-3">
                            <TabsTrigger value="general">
                                General Information
                            </TabsTrigger>
                            <TabsTrigger value="password">
                                Change Password
                            </TabsTrigger>
                            <TabsTrigger value="apikey">API Key</TabsTrigger>
                        </TabsList>
                        <TabsContent value="general">
                            <GeneralInfoTab user={user} updateUser={setUser}/>
                        </TabsContent>
                        <TabsContent value="password">
                            <ChangePasswordTab
                                userId={user ? user.id : undefined}
                            />
                        </TabsContent>
                        <TabsContent value="apikey">
                            <ApiKeyTab user={user} updateUser={setUser} />
                        </TabsContent>
                    </Tabs>
                </CardContent>
            </Card>
        </div>
    );
}
