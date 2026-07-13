declare module 'weapp-qrcode' {
  interface DrawQrcodeOptions {
    width: number;
    height: number;
    canvasId: string;
    text: string;
    typeNumber?: number;
    correctLevel?: number;
    background?: string;
    foreground?: string;
    x?: number;
    y?: number;
    callback?: (err: any) => void;
  }
  function drawQrcode(options: DrawQrcodeOptions): void;
  export default drawQrcode;
}
