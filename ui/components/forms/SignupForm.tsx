"use client";

import { IconEmail } from "@/components/icons/email";
import { IconEye } from "@/components/icons/eye";
import { IconEyeOff } from "@/components/icons/eye-off";
import { IconPasswordOutline } from "@/components/icons/password";
import { IconUserOutline } from "@/components/icons/user";
import { SignUp } from "@/lib/auth";
import { Spinner } from "@nextui-org/spinner";
import { set } from "date-fns";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { toast } from "react-toastify";

export default function SignUpForm() {
    const [email, setEmail] = useState("");
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [hidingPassword, setHidingPassword] = useState(true);
    const [confirmPassword, setConfirmPassword] = useState("");
    const [hidingConfirmPassword, setHidingConfirmPassword] = useState(true);
    const [isLoading, setIsLoading] = useState(false);
    const router = useRouter();
    const [errors, setErrors] = useState({
        email: "",
        password: "",
        confirmPassword: "",
        username: "",
    });

    const validate = () => {
        const newErrors = {
            email: "",
            password: "",
            confirmPassword: "",
            username: "",
        };

        if (!email) {
            newErrors.email = "Email is required";
        } else if (!/\S+@\S+\.\S+/.test(email)) {
            newErrors.email = "Email is invalid";
        }

        if (!username) {
            newErrors.username = "Username is required";
        } else if (username.length < 6) {
            newErrors.username = "Username must be >= 6 characters";
        } else if (username.length > 20) {
            newErrors.username = "Username must be <= 20 characters";
        }

        if (!password) {
            newErrors.password = "Password is required";
        } else if (password.length < 8) {
            newErrors.password = "Password must be at least 8 characters";
        }

        if (!confirmPassword) {
            newErrors.confirmPassword = "Please confirm your password";
        } else if (confirmPassword !== password) {
            newErrors.confirmPassword = "Passwords do not match";
        }

        setErrors(newErrors);

        return (
            !newErrors.email &&
            !newErrors.password &&
            !newErrors.confirmPassword
        );
    };

    const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        e.stopPropagation();
        setIsLoading(true);

        if (validate()) {
            const { fetchError } = await SignUp({email, username, password});
            if (fetchError) {
                toast.error(fetchError.message);
            } else {
                toast.success("Account created successfully");
                router.push("/")
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
            {errors.email && (
                <p className="text-red-500 text-xs font-semibold">
                    {errors.email}
                </p>
            )}

            <div className="flex px-4 gap-4 items-center rounded-md border border-gray-200 mt-4">
                <label htmlFor="username">
                    <IconUserOutline className="opacity-40 scale-125" />
                </label>
                <input
                    id="username"
                    aria-label="Username"
                    className="h-[50px] focus:outline-none flex-1"
                    placeholder="Username"
                    type="text"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                />
            </div>
            {errors.username && (
                <p className="text-red-500 text-xs font-semibold">
                    {errors.username}
                </p>
            )}

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
            {errors.password && (
                <p className="text-red-500 text-xs font-semibold">
                    {errors.password}
                </p>
            )}

            <div className="flex px-4 gap-4 items-center rounded-md border border-gray-200 mt-4">
                <label htmlFor="confirm-password">
                    <IconPasswordOutline className="opacity-40 scale-125" />
                </label>
                <input
                    id="confirm-password"
                    aria-label="Confirm Password"
                    className="h-[50px] focus:outline-none flex-1"
                    placeholder="Confirm Password"
                    type={hidingConfirmPassword ? "password" : "text"}
                    value={confirmPassword}
                    onChange={(e) => setConfirmPassword(e.target.value)}
                />
                {!hidingConfirmPassword ? (
                    <IconEye
                        className="scale-150 opacity-50"
                        onClick={() => setHidingConfirmPassword(true)}
                    />
                ) : (
                    <IconEyeOff
                        className="scale-150 opacity-50"
                        onClick={() => setHidingConfirmPassword(false)}
                    />
                )}
            </div>
            {errors.confirmPassword && (
                <p className="text-red-500 text-xs font-semibold">
                    {errors.confirmPassword}
                </p>
            )}

            <button
                type="submit"
                disabled={isLoading}
                className="w-full rounded-md flex justify-center items-center bg-blue-400 hover:bg-blue-500 text-white h-[50px] border-transparent border mt-4 font-semibold"
            >
                SIGN UP
                {isLoading && <Spinner className="ml-2" />}
            </button>
        </form>
    );
}
