import React from 'react';
import './Board.css';

class NewPostForm extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            apiUrl: this.props.apiUrl,
            email: 'noko',
            comment: '',
            subject: '',
            filename: '',
            threadNo: this.props.threadNo || null,
            image: null,
            error: null
        };
        this.fileInput = React.createRef();

        this.handleInputChange = this.handleInputChange.bind(this);
        this.handleFileChange = this.handleFileChange.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);
    }

    handleInputChange(event) {
        const target = event.target;
        const value = target.type === 'checkbox' ? target.checked : target.value;
        const name = target.name;

        this.setState({
            [name]: value
        });
    }

    handleFileChange(event) {
        this.setState({image: event.target.files[0]})
    }

    handleSubmit(event) {
        event.preventDefault();
        this.uploadForm()
    }

    uploadForm() {
        if (this.state.image === null && this.state.threadNo === null) {
            this.setState({error: "Image required!"});
            return;
        }

        let data = new FormData();
        data.append("email", this.state.email);
        data.append("comment", this.state.comment);

        if (this.state.image !== null) {
            data.append("filename", this.state.image.name);
            data.append("image", this.state.image);
        }

        let apiURL = this.state.apiUrl;
        if (this.state.threadNo === null) {
            apiURL = apiURL + '/thread';
            data.append("subject", this.state.subject)
        } else {
            apiURL = apiURL + '/post';
            data.append("threadNo", this.state.threadNo)
        }

        fetch(apiURL, {
            method: 'POST',
            body: data,
        }).then((res) => {
            if (res.ok) {
                console.log("Post Success!");
                this.setState({});
                window.location.reload()
            } else {
                console.log(res.status + " " + res.statusText);
            }
        }).catch(console.log);
    }

    showError() {
        return (
            <div className="error">{this.state.error}</div>
        )
    }

    render() {
        return (
            <div className="reply">
                {this.showError()}
                <form onSubmit={this.handleSubmit}>
                    <div className="field"><label>
                        Email:
                        <input type="text" name="email" value={this.state.value} onChange={this.handleInputChange}/>
                    </label></div>
                    <div className="field">
                        <label> Comment: <textarea name="comment" value={this.state.comment} onChange={this.handleInputChange}/> </label>
                    </div>
                    <div className="field">
                        <label htmlFor="Image">Image:</label>
                        <input type="file"
                               ref={this.fileInput}
                               onChange={this.handleFileChange}
                               id="image" name="image"
                               accept="image/png, image/jpeg"/>
                    </div>
                    <div className="field">
                        <input type="submit" value="Submit"/>
                    </div>
                </form>
            </div>
        );
    }
}

export default NewPostForm
