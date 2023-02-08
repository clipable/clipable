declare module "shaka-player-react" {
  import { Component } from "react";

  interface Props {
    autoPlay?: boolean;
    src: string;
  }

  export default class ShakaPlayer extends Component<Props> {}
}
