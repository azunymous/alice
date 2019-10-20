import React from 'react';
import {Link} from "react-router-dom";

import './Board.css';
import './Hover.css';


class Thread extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            status: "SUCCESS",
            board: this.props.board,
            no: this.props.no,
            thread: this.props.thread,
            limit: this.props.limit
        }
    }

    componentDidMount() {
        if (this.state.thread === undefined || this.state.thread === null || this.state.status === "FAILURE")
            this.getThread()
    }

    getThread() {
        fetch(process.env.REACT_APP_API_URL + '/thread?no=' + this.state.no)
            .then((response) => {
                return response.json()
            })
            .then((json) => {
                console.log(json);
                this.setState({
                    status: json.status,
                    thread: json.thread
                })
            }).catch(console.log);
    }

    displayImage(post) {
        if (process.env.REACT_APP_IMAGE_CONTEXT.startsWith("http")) {
            return process.env.REACT_APP_IMAGE_CONTEXT + "/" + post.image
        }
        return process.env.PUBLIC_URL + process.env.REACT_APP_IMAGE_CONTEXT + post.image;
    }

    displayThread(thread) {
        return (
            <div key={thread.post.no}>
                <hr/>
                <div className="thread">
                    <span className="image"><img alt={thread.post.filename}
                                                 src={this.displayImage(thread.post)}/></span><span
                    className="threadHeader">{thread.subject} <span
                    className="postName">{thread.post.name}</span> {thread.post.timestamp} No. <Link
                    to={"/" + this.state.board + "/res/" + thread.post.no}>{thread.post.no}</Link> <span className="quotedBy">{this.quotedBy(thread.post, thread)}</span></span>

                    <div><span className="content">{this.displayComment(thread.post)}</span></div>
                </div>
                <div className="replies">
                    {this.displayReplies(thread, 5)}
                </div>
            </div>
        );
    }

    displayReplies(thread) {
        if (this.state.limit !== null) {
            thread.replies = thread.replies.slice(-this.state.limit)
        }
        return thread.replies.map((post) => {
            return (
                this.displayPost(post, thread)
            )
        })
    }

    displayPost(post, thread, hover=false) {
        return <div key={post.no} className="post">
            {this.optionalImage(post)}
            <span className="postHeader"><span
                className="postName">{post.name}</span> {post.timestamp} No. {post.no} <span
                className="quotedBy">{this.quotedBy(post, thread, hover)}</span></span>
            <div><span className="content">{this.displayComment(post)}</span></div>
        </div>;
    }

    displayComment(post) {
        if (post.comment_segments == null) {
            return post.comment;
        }

        function formatAsClasses(segment) {
            if (segment === null || segment.format === null) {
                return "";
            }
            return segment.map((format) => {
                return format + " "
            })
        }

        return post.comment_segments.map((segment, i) => {
            return (
                <div className={formatAsClasses(segment.format)} key={i}>{segment.segment}<br/></div>
            )
        })
    }


    quotedBy(post, thread, hover=false) {
        const Hover = ({ onHover, children }) => (
            <span className="hover">
                <span className="hover__no-hover">{children}</span>
                <span className="hover__hover">{onHover}</span>
            </span>
        );


        if (post.quoted_by == null || hover) {
            return <span/>
        }
        return post.quoted_by.map((postNo, i) => {
            return (
                <Hover key={i} onHover={this.displayPostHover(this.findPost(thread, postNo), thread)}>
                    <span className="noQuote" key={i}>>>{postNo} </span>
                </Hover>

            )
        })
    }

    displayPostHover(post, thread) {
        if (post === null) {
            return <span className="hoveredPost"/>
        }

        return this.displayPost(post, thread, true)
    }

    optionalImage(post) {
        if (post.image != null && post.image !== "") {
            return <span className="image"><img src={this.displayImage(post)}
                                                alt={post.filename}/></span>
        } else {
            return <span className="noImage"/>
        }
    }

    render() {
        if (this.state.thread === undefined || this.state.thread === null || this.state.status === "FAILURE") {
            return (<div>. . .</div>)
        }
        return this.displayThread(this.state.thread)
    }

    findPost(thread, postNo) {
        if (thread.post.no === postNo) {
            return thread.post
        }

        for (let i = 0; i < thread.replies.length; i++) {
            if (thread.replies[i].no === postNo) {
                return thread.replies[i]
            }
        }
        return null;
    }
}

export default Thread