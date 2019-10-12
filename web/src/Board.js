import React from 'react';
import './Board.css';
import NewThreadForm from "./NewThreadForm";

class Board extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            threads: []
        }
    }


    getAllThreads() {
        fetch('/thread/all/')
            .then((response) => {
                return response.json()
            })
            .then((json) => {
                console.log(json)
                this.setState({
                    threads: json
                })
            }).catch(console.log);
    }

    componentDidMount() {
        this.getAllThreads();
    }

    render() {
        return (
            <div>
                <h2>{this.props.name}</h2>
                <NewThreadForm/>
                {this.allThreads()}
            </div>
        );
    }

    allThreads() {
        if (this.state.threads == null) {
            return (<div> . . . </div>)
        }
        let threads = this.state.threads.slice(0, 10)
        return threads.map((thread) => {
            return this.displayThread(thread);
        })
    }

    displayThread(thread) {
        return (
            <div key={thread.post.no}>
                <hr/>
                <div className="thread">
                    <span className="image"><img alt={thread.post.filename} src={process.env.PUBLIC_URL + "/images/" + thread.post.image}/></span><span className="threadHeader">{thread.subject} <span
                    className="postName">{thread.post.name}</span> {thread.post.timestamp} No. {thread.post.no}</span>

                    <div><span className="content">{thread.post.comment}</span></div>
                </div>
                <div className="replies">
                    {this.displayReplies(thread, 5)}
                </div>
            </div>
        )
    }

    displayReplies(thread, limit = null) {
        if (limit !== null) {
            thread.replies = thread.replies.slice(-limit)
        }

        function optionalImage(post) {
            if (post.image != null && post.image !== "") {
                return <span className="image"><img src={process.env.PUBLIC_URL + "/images/" + post.image} alt={post.filename}/></span>
            } else {
                return <span className="noImage"/>
            }
        }

        return thread.replies.map((post) => {
            return (
                <div key={post.no} className="post">
                    {optionalImage(post)}
                    <span className="postHeader"><span
                        className="postName">{post.name}</span> {post.timestamp} No. {post.no}</span>
                    <div><span className="content">{post.comment}</span></div>
                </div>
            )
        })
    }
}

export default Board