import { AtomEffect } from "recoil";

export const localStorageEffect: AtomEffect<any> = ({ onSet, setSelf, node }) => {
  const storedData = localStorage.getItem(node.key);
  if (storedData != null) {
    setSelf(JSON.parse(storedData));
  }

  onSet((newIds, _, isReset) => {
    isReset
      ? localStorage.removeItem(node.key)
      : localStorage.setItem(node.key, JSON.stringify(newIds));
  });
};
