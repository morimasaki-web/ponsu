import { useEffect, useState } from "react";

/**
 * デバウンス処理を行うカスタムフック
 * @param value デバウンス対象の値
 * @param delay 遅延時間（ミリ秒）
 * @returns デバウンスされた値
 */
export function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value);

  useEffect(() => {
    // 指定時間後に値を更新するタイマーをセット
    const timer = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    // クリーンアップ: 値が変更されたらタイマーをキャンセル
    return () => {
      clearTimeout(timer);
    };
  }, [value, delay]);

  return debouncedValue;
}
