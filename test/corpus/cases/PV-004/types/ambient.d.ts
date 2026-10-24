declare const process: {
  stdout: { write(chunk: string): void };
  stderr: { write(chunk: string): void };
  exit(code: number): never;
};

declare module "commonmark" {
  interface WalkerEvent {
    entering: boolean;
    node: CmNode;
  }

  interface Walker {
    next(): WalkerEvent | null;
  }

  interface CmNode {
    type: string;
    literal: string | null;
    info: string | null;
    _isFenced: boolean;
    walker(): Walker;
  }

  export class Parser {
    parse(input: string): CmNode;
  }
}
