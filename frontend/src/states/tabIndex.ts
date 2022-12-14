import { atom } from "recoil";
import { localStorageEffect } from "./effects/localStorageEffect";

export const tabIndexState = atom<number>({
  key: "tabIndexState",
  default: 0,
  effects: [localStorageEffect],
});
