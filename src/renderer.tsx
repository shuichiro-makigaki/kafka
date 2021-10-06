import './index.css';

import React, {useState, useEffect} from 'react';
import ReactDOM from 'react-dom';
import InfiniteScroll from 'react-infinite-scroll-component';
import {Line} from 'rc-progress';
import {default as axios} from 'axios';
import {ContextBridgeApi} from './preload';
import {MovieModel as MovieModel} from './models';
import * as path from "path";
import {DateTime} from "luxon";
import shuffle from "shuffle-array";

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore
const IPC: ContextBridgeApi = window.IPC;

function Thumbnails(props: {movie: MovieModel, thumbnailCount: number}) {
    const {movie} = props;

    return (
        <div className="thumbnails">
            {
                [...Array(props.thumbnailCount)].map((e, i) =>
                    <img src={`http://localhost:8080/movie/${movie.id}/thumbnail/${i}`} alt="" key={i} />)
            }
        </div>
    );
}

function Movie(props: {movie: MovieModel, thumbnailCount: number}) {
    const {movie} = props;

    return (
        <div className="movie"
             onDoubleClick={e => IPC.openMovie(movie)}
             onContextMenu={e => IPC.openMovieContextMenu(movie)}
        >
            <Thumbnails movie={movie} thumbnailCount={props.thumbnailCount} />
            <div className="title">
                {movie.title}
            </div>
        </div>
    );
}

function Movies(props: {movies: MovieModel[], sortType: string, thumbnailCount: number}) {
    const [movieList, updateMovieList] = useState<MovieModel[]>([]);
    const [dataLength, updateDataLength] = useState<number>(100);
    const [delTargets, updateDelTargets] = useState<string[]>([]);

    useEffect(() => {
        if (dataLength == 0) {
            updateDataLength(100);
        }
    }, [dataLength]);

    useEffect(() => {
        window.scrollTo(0, 0);
        switch (props.sortType) {
            case 'last_modified_time':
                updateMovieList(
                    movieList.sort((a:MovieModel, b:MovieModel) => {
                        return a.lastModifiedTime < b.lastModifiedTime ? 1 : a.lastModifiedTime > b.lastModifiedTime ? -1 : 0;
                    })
                );
                break;
            case 'random':
                updateMovieList(shuffle(movieList));
                break;
        }
        updateDataLength(0);
    }, [props.sortType]);

    useEffect(() => {
        updateMovieList(props.movies);
    }, [props.movies]);

    useEffect(() => {
        const promises:Promise<any>[] = [];
        delTargets.forEach(target => {
            promises.push(
                axios.delete(`http://localhost:8080/movie/${target}`).then(resp => {
                    movieList.splice(movieList.findIndex(e => e.id === target), 1);
                })
            );
        });
        Promise.all(promises).then(values => {
            updateMovieList(movieList);
            updateDataLength(dataLength+delTargets.length);
        });
    }, [delTargets]);

    useEffect(() => {
        IPC.onDeleteMovie(movieId => {
            updateDelTargets([movieId]);
        });
        IPC.onGenerateThumbnails(movieId => {
            axios.post(`http://localhost:8080/movie/${movieId}/thumbnail`, {thumbnailCount: props.thumbnailCount})
                .catch(reason => {
                    console.error(reason);
                });
        });
    }, []);

    return (
        <InfiniteScroll
            dataLength={dataLength}
            next={() => {updateDataLength(dataLength+5)}}
            hasMore={dataLength != movieList.length}
            loader={<div>Loading...</div>}
            className="movies"
            scrollThreshold="100%"
        >
            {movieList.slice(0, dataLength).map(v => <Movie movie={v} thumbnailCount={props.thumbnailCount} key={v.id} />)}
        </InfiniteScroll>
    );

}

function App() {
    const [movieList, updateMovieList] = useState<MovieModel[]>([]);
    const [sortType, updateSortType] = useState<string>(null);
    const [thumbnailCount, updateThumbnailCount] = useState<number>(0);
    const [serverPort, updateServerPort] = useState<number>(8080);

    useEffect(() => {
        IPC.onSendSortType(sortType => {
            // if rand updateSortType(null); for force render
            updateSortType(sortType);
        });
        IPC.onSendThumbnailCount(thumbnailCount => {
            updateThumbnailCount(thumbnailCount);
        });
        axios.get(`http://localhost:${serverPort}/movie`).then(resp => {
            const movies = resp.data.map((v: any) => {
                const m = new MovieModel();
                m.id = v.id;
                m.file = v.path;
                m.title = path.basename(m.file.replaceAll('\\', '/'));
                m.lastModifiedTime = DateTime.fromISO(v.last_modified_time);
                return m;
            });
            IPC.getRecentSortType().then(sortType => {
                switch (sortType) {
                    case 'last_modified_time':
                        movies.sort((a:MovieModel, b:MovieModel) => a.lastModifiedTime<b.lastModifiedTime?1:a.lastModifiedTime>b.lastModifiedTime?-1:0);
                        break;
                    case 'random':
                        shuffle(movies);
                        break;
                }
                updateMovieList(movies);
            });
        });
        IPC.getRecentThumbnailCount().then(v => {
            updateThumbnailCount(v);
        });
    }, []);

    return <Movies movies={movieList} sortType={sortType} thumbnailCount={thumbnailCount}/>;
}

function Footer() {
    const [currentPercent, updateCurrentPercent] = useState(0);
    const [message, updateMessage] = useState('');

    useEffect(() => {
        IPC.onSendPercent((percent: number) => {
            updateCurrentPercent(percent);
        });
        IPC.onSendMessage((message: string) => {
            console.log(message);
            updateMessage(message);
        });
    }, []);

    return (
        <>
            <div className="message">{message}</div>
            <div className="progress-bar">
                <Line className="line" percent={currentPercent} strokeWidth={6} trailWidth={6} />
            </div>
            <div className="progress-value">{Math.round(currentPercent)}%</div>
        </>
    );
}

ReactDOM.render(<App />, document.getElementById('app'));
ReactDOM.render(<Footer />, document.getElementById('footer'));
