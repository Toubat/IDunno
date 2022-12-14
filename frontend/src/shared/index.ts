export const STAT_SERVER_ADDRESS = "http://fa22-cs425-2401.cs.illinois.edu:7000";

/**
 * Given a Date object, return a string  that display one of
 * the following formats:
 * - "xx seconds ago"
 * - "xx minutes ago"
 * - "xx hours ago"
 * - "xx days ago"
 * - "xx weeks ago"
 * - "xx months ago"
 * - "xx years ago"
 * @param date
 */
export const formatDate = (dateInUTC: Date) => {
  // subtract 6 hours from date
  const date = new Date(dateInUTC.getTime() - 6 * 60 * 60 * 1000);
  const seconds = Math.floor((new Date().getTime() - date.getTime()) / 1000);

  let interval = Math.floor(seconds / 31536000);
  if (interval > 1) {
    return interval + " years ago";
  }

  interval = Math.floor(seconds / 2592000);
  if (interval > 1) {
    return interval + " months ago";
  }

  interval = Math.floor(seconds / 604800);
  if (interval > 1) {
    return interval + " weeks ago";
  }

  interval = Math.floor(seconds / 86400);
  if (interval > 1) {
    return interval + " days ago";
  }

  interval = Math.floor(seconds / 3600);
  if (interval > 1) {
    return interval + " hours ago";
  }

  interval = Math.floor(seconds / 60);
  if (interval > 1) {
    return interval + " minutes ago";
  }

  return Math.floor(seconds) + " seconds ago";
};

export const toDecimalPlace = (num: number, decimalPlace: number) => {
  const multiplier = Math.pow(10, decimalPlace);
  return Math.round(num * multiplier) / multiplier;
};
