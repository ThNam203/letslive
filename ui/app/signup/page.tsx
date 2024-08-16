'use client'

import { CarbonEmail } from "@/components/icons/email";
import { LogosFacebook } from "@/components/icons/facebook";
import { LogosGoogleIcon } from "@/components/icons/google";
import { MdiPasswordOutline } from "@/components/icons/password";

export default function SignUpPage() {
  return (
    <section className="flex items-center justify-center h-screen w-screen">
      <div className="flex flex-col justify-center rounded-xl p-12 bg-white">
        <h1 className="text-lg font-bold">LET'S LIVE</h1>
        <h1 className="text-2xl font-bold mb-2">
          Welcome! Sign up for a new world?
        </h1>
        <p className="text-md">Choose a method below to begin</p>
        <div className="flex gap-2 my-2">
          <div onClick={() => {
            window.location.href = "http://localhost:8000/v1/auth/google"
        }
            } className="flex items-center justify-center gap-1 rounded-md border-gray-300 border w-[200px] h-[50px] hover:cursor-pointer">
            <LogosGoogleIcon /> Google
          </div>
          <div className="flex items-center justify-center gap-1 rounded-md border-gray-300 border w-[200px] h-[50px] hover:cursor-pointer">
            <LogosFacebook /> Facebook
          </div>
        </div>
        <div className="flex items-center justify-center w-full mt-2 mb-4">
          <hr className="bg-gray-400 h-[2px] flex-1" />
          <p className="text-center mx-4 text-gray-500">or</p>
          <hr className="bg-gray-400 h-[2px] flex-1" />
        </div>
        <form>
          <div className="flex px-4 gap-4 items-center rounded-md border border-gray-200 mb-4">
            <CarbonEmail className="opacity-40 scale-125" />
            <input
              id="username"
              className="h-[50px] focus:outline-none w-full"
              placeholder="Email"
            ></input>
          </div>
          <div className="flex px-4 gap-4 items-center rounded-md border border-gray-200">
            <MdiPasswordOutline className="opacity-40 scale-125" />
            <input
              id="password"
              className="h-[50px] focus:outline-none w-full"
              placeholder="Password"
              type="password"
            ></input>
          </div>
          <button className="w-full rounded-md flex justify-center items-center bg-blue-400 hover:bg-blue-500 text-white h-[50px] border-transparent border mt-4 font-semibold">
            SIGN UP
          </button>
        </form>
        <p className="text-end text-sm opacity-80 mt-4">
          Already have an account?
          <a
            href="/signin"
            className="ml-2 text-blue-400 font-bold hover:opacity-100"
          >
            Sign in
          </a>
        </p>
      </div>
    </section>
  );
}
