declare module "shaka-player-react" {
  import { Component } from "react";

  interface WithRef {
    ref: MutableRefObject<any>;
    chromeless?: boolean;
  }

  interface WithVideoSrc {
    autoPlay?: boolean;
    src: string;
    chromeless?: boolean;
  }

  type Props = WithRef | WithVideoSrc

  export default class ShakaPlayer extends Component<Props> { }
}
