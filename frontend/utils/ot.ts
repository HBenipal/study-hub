// ot helper functions
export const applyOperation = (text: string, operation: any) => {
  if (operation.type === "insert") {
    return (
      text.slice(0, operation.position) +
      operation.text +
      text.slice(operation.position)
    );
  } else if (operation.type === "delete") {
    return (
      text.slice(0, operation.position) +
      text.slice(operation.position + operation.length)
    );
  }
  return text;
};

export const transformCursor = (cursor: number, operation: any) => {
  if (operation.type === "insert") {
    if (cursor >= operation.position) {
      return cursor + operation.text.length;
    }
  } else if (operation.type === "delete") {
    if (cursor > operation.position + operation.length) {
      return cursor - operation.length;
    } else if (cursor > operation.position) {
      return operation.position;
    }
  }
  return cursor;
};

export const computeOperations = (oldText: string, newText: string) => {
  if (oldText == null || newText == null) return [];
  let i = 0;
  while (i < oldText.length && i < newText.length && oldText[i] === newText[i])
    i++;
  let oldEnd = oldText.length;
  let newEnd = newText.length;
  while (
    oldEnd > i &&
    newEnd > i &&
    oldText[oldEnd - 1] === newText[newEnd - 1]
  ) {
    oldEnd--;
    newEnd--;
  }
  const operations = [];
  if (oldEnd > i) {
    operations.push({ type: "delete", position: i, length: oldEnd - i });
  }
  if (newEnd > i) {
    operations.push({
      type: "insert",
      position: i,
      text: newText.slice(i, newEnd),
    });
  }
  return operations;
};
