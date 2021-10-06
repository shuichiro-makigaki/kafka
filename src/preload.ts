import {MovieModel} from './models';
import {contextBridge, ipcRenderer, IpcRendererEvent} from "electron";

export class ContextBridgeApi {
    public static readonly API_KEY = 'IPC';
    
    public openMovie = (movie: MovieModel) => {
        return ipcRenderer.invoke('openMovie', movie);
    };

    public getRecentSortType = () => {
        return ipcRenderer.invoke('getRecentSortType');
    }

    public getRecentThumbnailCount = () => {
        return ipcRenderer.invoke('getRecentThumbnailCount');
    }

    public openMovieContextMenu = (movie: MovieModel) => {
        return ipcRenderer.invoke('openMovieContextMenu', movie);
    }

    public onDeleteMovie = (rendererListener: (movieId: string) => void) => {
        ipcRenderer.on(
            'deleteMovie',
            (event: IpcRendererEvent, movieId: string) => {
                rendererListener(movieId);
            }
        );
    }

    public onGenerateThumbnails = (rendererListener: (movieId: string) => void) => {
        ipcRenderer.on(
            'generateThumbnails',
            (event: IpcRendererEvent, movieId: string) => {
                rendererListener(movieId);
            }
        );
    }

    public onSendSortType = (rendererListener: (sortType: string) => void) => {
        ipcRenderer.on(
            'sendSortType',
            (event: IpcRendererEvent, sortType: string) => {
                rendererListener(sortType);
            }
        );
    };

    public onSendThumbnailCount = (rendererListener: (thumbnailCount: number) => void) => {
        ipcRenderer.on(
            'sendThumbnailCount',
            (event: IpcRendererEvent, thumbnailCount: number) => {
                rendererListener(thumbnailCount);
            }
        );
    };

    public onSendMessage = (rendererListener: (message: string) => void) => {
        ipcRenderer.on(
            'sendMessage',
            (event: IpcRendererEvent, message: string) => {
                rendererListener(message);
            }
        );
    };

    public onSendPercent = (rendererListener: (percent: number) => void) => {
        ipcRenderer.on(
            'sendPercent',
            (event: IpcRendererEvent, percent: number) => {
                rendererListener(percent);
            }
        );
    };
}

contextBridge.exposeInMainWorld(ContextBridgeApi.API_KEY, new ContextBridgeApi());


