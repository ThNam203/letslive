"use client";

import FormErrorText from "@/components/forms/FormErrorText";
import { IconEmail } from "@/components/icons/email";
import { IconEye } from "@/components/icons/eye";
import { IconEyeOff } from "@/components/icons/eye-off";
import { IconPasswordOutline } from "@/components/icons/password";
import GLOBAL from "@/global";
import { LogIn } from "@/lib/api/auth";
import { Spinner } from "@nextui-org/spinner";
import { set } from "date-fns";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { toast } from "react-toastify";

export default function LogInForm() {
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [hidingPassword, setHidingPassword] = useState(true);
    const [isLoading, setIsLoading] = useState(false);
    const router = useRouter();
    const [errors, setErrors] = useState({
        email: "",
        password: "",
    });

    const validate = () => {
        const newErrors = { email: "", password: "" };

        if (!email) {
            newErrors.email = "Email is required";
        } else if (!/\S+@\S+\.\S+/.test(email)) {
            newErrors.email = "Email is invalid";
        }

        if (!password) {
            newErrors.password = "Password is required";
        } else if (password.length < 8) {
            newErrors.password = "Password must be at least 8 characters";
        }

        setErrors(newErrors);

        return !newErrors.email && !newErrors.password;
    };

    const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        e.stopPropagation();
        setIsLoading(true);

        if (validate()) {
            const { fetchError } = await LogIn({ email, password });
            if (fetchError) {
                toast.error(fetchError.message);
            } else {
                router.push("/");
            }
        }
        setIsLoading(false);
    };

    return (
        <form onSubmit={handleSubmit}>
            <div className="flex px-4 gap-4 items-center rounded-md border border-gray-200">
                <label htmlFor="email">
                    <IconEmail className="opacity-40 scale-125" />
                </label>
                <input
                    id="email"
                    aria-label="Email"
                    className="h-[50px] focus:outline-none flex-1"
                    placeholder="Email"
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                />
            </div>
            <FormErrorText textError={errors.email} />
            <div className="flex px-4 gap-4 items-center rounded-md border border-gray-200 mt-4">
                <label htmlFor="password">
                    <IconPasswordOutline className="opacity-40 scale-125" />
                </label>
                <input
                    id="password"
                    aria-label="Password"
                    className="h-[50px] focus:outline-none flex-1"
                    placeholder="Password"
                    type={hidingPassword ? "password" : "text"}
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                />
                {!hidingPassword ? (
                    <IconEye
                        className="scale-150 opacity-50"
                        onClick={() => setHidingPassword(true)}
                    />
                ) : (
                    <IconEyeOff
                        className="scale-150 opacity-50"
                        onClick={() => setHidingPassword(false)}
                    />
                )}
            </div>
            <FormErrorText textError={errors.password} />
            <button
                type="submit"
                disabled={isLoading}
                className="w-full rounded-md flex justify-center items-center bg-blue-400 hover:bg-blue-500 text-white h-[50px] border-transparent border mt-4 font-semibold"
            >
                LOG IN
                {isLoading && <Spinner className="ml-2" />}
            </button>
        </form>
    );
}
