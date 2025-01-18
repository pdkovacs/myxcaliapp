import { isEqual } from "lodash";
import { useEffect, useMemo, useRef, useState } from "react";
import { LinearProgress } from "@mui/material";
import "./App.css";
import { Excalidraw, MainMenu } from "@excalidraw/excalidraw";
import { OpenDrawingDialog } from "./features/drawing/OpenDrawing";
import type { ExcalidrawImperativeAPI } from "@excalidraw/excalidraw/types/types";
import { useAppDispatch, useAppSelector } from "./app/hooks";
import { selectSavedDrawing, selectDrawingToEditStatus, selectCurrentDrawingContent, drawingContentChanged } from "./features/drawing/drawingSlice";
import { SaveDrawingDialog } from "./features/drawing/SaveDrawing";

const App = () => {

  const [excalidrawAPI, setExcalidrawAPI] = useState<ExcalidrawImperativeAPI | null>(null);
  const excalidrawAPIUnsubscribe = useRef<(() => void) | null>(null);

  const currentDrawingStatus = useAppSelector(selectDrawingToEditStatus);
  const savedDrawing = useAppSelector(selectSavedDrawing);
  const currentContent = useAppSelector(selectCurrentDrawingContent);

  const [openDrawingDialogOpen, setOpenDrawingDialogOpen] = useState(false);
  const [saveDrawingDialogOpen, setSaveDrawingDialogOpen] = useState(false);

  const dispatch = useAppDispatch();

  useEffect(() => {
    if (excalidrawAPI) {
      if (excalidrawAPIUnsubscribe.current !== null) {
        excalidrawAPIUnsubscribe.current();
      }
      excalidrawAPIUnsubscribe.current = excalidrawAPI.onChange(() => {
        dispatch(drawingContentChanged(JSON.stringify(excalidrawAPI.getSceneElements())));
      });
    }
  }, [excalidrawAPI, savedDrawing]);

  useEffect(() => {
    const sceneData = {
      elements: savedDrawing.content ? JSON.parse(savedDrawing.content) : [],
      appState: {}
    };

    excalidrawAPI?.updateScene(sceneData);
  }, [savedDrawing]);

  const contentHasChanged = useMemo(() => {
    return !isEqual(savedDrawing.content, currentContent);
  }, [savedDrawing, currentContent]);

  console.log(">>>>>>>> contentHasChanged", contentHasChanged);

  return (
    <div className="App">
      <h1 style={{ textAlign: "center" }}>My Excalidraw App</h1>
      {currentDrawingStatus === "loading" && <LinearProgress sx={{ marginTop: "-4px" }} />}
      <div style={{ height: "calc(100vh - 80px)" }}>
        <Excalidraw excalidrawAPI={api => setExcalidrawAPI(api)}>
          <MainMenu>
            <MainMenu.Item onSelect={() => setOpenDrawingDialogOpen(true)}>
              Open
            </MainMenu.Item>
            <MainMenu.Item disabled={!contentHasChanged} onSelect={() => setSaveDrawingDialogOpen(true)}>
              Save
            </MainMenu.Item>
          </MainMenu>
        </Excalidraw>
      </div>
      <OpenDrawingDialog open={openDrawingDialogOpen} onClose={() => setOpenDrawingDialogOpen(false)} />
      <SaveDrawingDialog open={saveDrawingDialogOpen} onClose={() => setSaveDrawingDialogOpen(false)} />
    </div>
  );
};

export default App;
